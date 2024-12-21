package mini

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/metafates/mangal/color"
	"github.com/metafates/mangal/style"
	"github.com/samber/lo"
)

var (
	yellowStyle = func(s string) string {
		truncated := style.Truncate(50)(s)
		return style.New().Width(50).Foreground(color.Yellow).Render(truncated)
	}

	cyanStyle = func(s string) string {
		truncated := style.Truncate(50)(s)
		return style.New().Width(50).Foreground(color.Cyan).Render(truncated)
	}

	redStyle = func(s string) string {
		truncated := style.Truncate(50)(s)
		return style.New().Width(50).Foreground(color.Red).Render(truncated)
	}

	colorize = map[string]func(string) string{
		"yellow": yellowStyle,
		"cyan":   cyanStyle,
		"red":    redStyle,
	}
)

func progress(msg string) (eraser func()) {
	msg = style.New().Foreground(color.Blue).Render(msg)
	fmt.Printf("\r%s", msg)

	return func() {
		fmt.Printf("\r%s\r", strings.Repeat(" ", len(msg)))
	}
}

func title(t string) {
	fmt.Println(style.Bold(t))
}

func fail(t string) {
	fmt.Println(style.New().Bold(true).Foreground(color.Red).Render(t))
}

func menu[T fmt.Stringer](items []T, options ...*bind) (*bind, T, error) {
	styles := map[int]func(string) string{
		0: yellowStyle,
		1: cyanStyle,
		2: redStyle,
	}

	for i, item := range items {
		s := fmt.Sprintf("(%d) %s", i+1, item.String())
		fmt.Println(styles[i%2](s))
	}

	options = append(options, quit)
	for i, option := range options {
		s := fmt.Sprintf("(%s) %s", option.A, option.B)
		s = style.Truncate(50)(s)

		if option == quit {
			fmt.Println(styles[2](s))
		} else {
			fmt.Println(styles[i%2](s))
		}
	}

	isValidOption := func(s string) bool {
		return lo.Contains(lo.Map(options, func(o *bind, _ int) string {
			return o.A
		}), s)
	}

	in, err := getInput(func(s string) bool {
		num, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return isValidOption(s)
		}
		return 0 < num && int(num-1) < len(items)+1
	})

	var t T

	if err != nil {
		return nil, t, err
	}

	if num, ok := in.asInt(); ok {
		return nil, items[num-1], nil
	}

	b, _ := parseBind(in.value)
	return b, t, nil
}
