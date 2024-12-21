package anilist

import "fmt"

// mangaSubquery common manga query used for getting manga by id or searching it by name
var mangaSubquery = `
id
idMal
title {
	romaji
	english
	native
}
description(asHtml: false)
tags {
	name
	description
	rank
	category
}
genres
coverImage {
	extraLarge
	large
	medium
	color
}
bannerImage
characters (page: 1, perPage: 25) {
	nodes {
		id
		name {
			full
			native
		}
		description
		image {
			large
		}
	}
}
startDate {
	year
	month	
	day
}
endDate {
	year
	month	
	day
}
staff {
	edges {
		role
		node {
			id
			name {
				full
				native
			}
			description
			image {
				large
			}
		}
	}
}
status
synonyms
siteUrl
chapters
countryOfOrigin
externalLinks {
	url
	site
	type
}
format
volumes
averageScore
popularity
meanScore
isLicensed
updatedAt
recommendations {
	nodes {
		mediaRecommendation {
			id
			title {
				romaji
				english
				native
			}
		}
	}
}
relations {
	edges {
		relationType
		node {
			id
			title {
				romaji
				english
				native
			}
		}
	}
}
`

// searchByNameQuery query used for searching manga by name
var searchByNameQuery = fmt.Sprintf(`
query ($query: String) {
	Page (page: 1, perPage: 30) {
		media (search: $query, type: MANGA, sort: [SEARCH_MATCH, POPULARITY_DESC]) {
			%s
		}
	}
}
`, mangaSubquery)

// searchByIDQuery query used for searching manga by id
var searchByIDQuery = fmt.Sprintf(`
query ($id: Int) {
	Media (id: $id, type: MANGA) {
		%s
	}
}`, mangaSubquery)
