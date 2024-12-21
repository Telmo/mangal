package anilist

import (
	"encoding/json"
	"fmt"
	"github.com/metafates/mangal/key"
	"github.com/metafates/mangal/log"
	"github.com/metafates/mangal/util"
	"github.com/metafates/mangal/where"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	levenshtein "github.com/ka-weihe/fast-levenshtein"
)

// normalizedName returns a normalized name for comparison
func normalizedName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// SetRelation sets the relation between a manga name and an anilist id
func SetRelation(name string, to *Manga) error {
	err := relationCacher.Set(name, to.ID)
	if err != nil {
		return err
	}

	if id := idCacher.Get(to.ID); id.IsAbsent() {
		return idCacher.Set(to.ID, to)
	}

	return nil
}

// FindClosest returns the closest manga to the given name.
// It will levenshtein compare the given name with all the manga names in the cache.
func FindClosest(name string) (*Manga, error) {
	name = normalizedName(name)
	
	// Try exact match first
	mangas, err := SearchByName(name)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Try to find an exact match
	for _, manga := range mangas {
		if normalizedName(manga.Name()) == name {
			log.Info("Found exact match: " + manga.Name())

			// Write debug file if debug flag is set
			if viper.GetBool(key.MetadataDebug) {
				debugDir := filepath.Join(where.Config(), "debug")
				debugFile := filepath.Join(debugDir, "anilist_exact_match.json")
				log.Infof("Writing exact match result to debug directory: %s", debugFile)
				if data, err := json.MarshalIndent(manga, "", "  "); err != nil {
					log.Errorf("Failed to marshal exact match result: %v", err)
				} else {
					if err := os.MkdirAll(debugDir, 0755); err != nil {
						log.Errorf("Failed to create debug directory: %v", err)
					} else {
						if err := os.WriteFile(debugFile, data, 0644); err != nil {
							log.Errorf("Failed to write exact match result: %v", err)
						} else {
							log.Infof("Successfully wrote exact match result to %s", debugFile)
						}
					}
				}
			}

			_ = relationCacher.Set(name, manga.ID)
			_ = idCacher.Set(manga.ID, manga)
			return manga, nil
		}
	}

	// Try alternative titles
	for _, manga := range mangas {
		for _, title := range manga.Synonyms {
			if normalizedName(title) == name {
				log.Info("Found match in alternative titles: " + manga.Name())
				_ = relationCacher.Set(name, manga.ID)
				_ = idCacher.Set(manga.ID, manga)
				return manga, nil
			}
		}
	}

	// Try fuzzy match
	if len(mangas) > 0 {
		closest := lo.MinBy(mangas, func(a, b *Manga) bool {
			return levenshtein.Distance(
				name,
				normalizedName(a.Name()),
			) < levenshtein.Distance(
				name,
				normalizedName(b.Name()),
			)
		})

		log.Info("Found closest match: " + closest.Name())
		_ = relationCacher.Set(name, closest.ID)
		_ = idCacher.Set(closest.ID, closest)
		return closest, nil
	}

	// Try with a simpler search
	words := strings.Split(name, " ")
	if len(words) > 1 {
		// Try with just the first two words
		simpleName := strings.Join(words[:util.Min(len(words), 2)], " ")
		log.Infof(`No results found for "%s", trying "%s"`, name, simpleName)
		return FindClosest(simpleName)
	}

	err = fmt.Errorf("no results found on Anilist for manga %s", name)
	log.Error(err)
	_ = relationCacher.Set(name, -1)
	return nil, err
}

// findClosest returns the closest manga to the given name.
// It will levenshtein compare the given name with all the manga names in the cache.
func findClosest(name, originalName string, try, limit int) (*Manga, error) {
	if try >= limit {
		err := fmt.Errorf("no results found on Anilist for manga %s", name)
		log.Error(err)
		_ = relationCacher.Set(originalName, -1)
		return nil, err
	}

	id := relationCacher.Get(name)
	if id.IsPresent() {
		if id.MustGet() == -1 {
			return nil, fmt.Errorf("no results found on Anilist for manga %s", name)
		}

		if manga, ok := idCacher.Get(id.MustGet()).Get(); ok {
			if try > 0 {
				_ = relationCacher.Set(originalName, manga.ID)
			}
			return manga, nil
		}
	}

	// search for manga on anilist
	mangas, err := SearchByName(name)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if id.IsPresent() {
		found, ok := lo.Find(mangas, func(item *Manga) bool {
			return item.ID == id.MustGet()
		})

		if ok {
			return found, nil
		}

		// there should be a manga with the id in the cache, but it wasn't found
		// this means that the manga was deleted from anilist
		// remove the id from the cache
		_ = relationCacher.Delete(name)
		log.Infof("Manga with id %d was deleted from Anilist", id.MustGet())
	}

	if len(mangas) == 0 {
		// try again with a different name
		words := strings.Split(name, " ")
		if len(words) <= 2 {
			// trigger limit, proceeding further will only make things worse
			return findClosest(name, originalName, limit, limit)
		}

		// one word less
		alternateName := strings.Join(words[:util.Max(len(words)-1, 1)], " ")
		log.Infof(`No results found on Anilist for manga "%s", trying "%s"`, name, alternateName)
		return findClosest(alternateName, originalName, try+1, limit)
	}

	// find the closest match
	closest := lo.MinBy(mangas, func(a, b *Manga) bool {
		return levenshtein.Distance(
			name,
			normalizedName(a.Name()),
		) < levenshtein.Distance(
			name,
			normalizedName(b.Name()),
		)
	})

	log.Info("Found closest match: " + closest.Name())

	save := func(n string) {
		if id := relationCacher.Get(n); id.IsAbsent() {
			_ = relationCacher.Set(n, closest.ID)
		}
	}

	save(name)
	save(originalName)

	_ = idCacher.Set(closest.ID, closest)
	return closest, nil
}
