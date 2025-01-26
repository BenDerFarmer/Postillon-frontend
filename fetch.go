package main

import (
	"strings"

	"github.com/gocolly/colly"
)

func fetch(url string) Post {

	c := colly.NewCollector()

	post := Post{}
	post.Images = make(map[uint8]Image)
	post.Extra = make(map[uint8]string)

	c.OnHTML(".entry-title", func(h *colly.HTMLElement) {
		if strings.ToLower(h.Text) != h.Text {
			post.Heading = h.Text
		}
	})

	c.OnHTML(".post-body", func(e *colly.HTMLElement) {

		var count uint8 = 0

		e.ForEach("*", func(i int, h *colly.HTMLElement) {
			if h.Name == "p" {
				count += 1
			}
			if h.Name == "noscript" {
				post.Extra[count-1] = h.Text
			}
			if h.Name == "img" && h.Attr("src") != "https://vg07.met.vgwort.de/na/6a3e6e06c7624bdc91bf0489eb9b722e" {
				image := Image{
					Source: strings.Replace(h.Attr("src"), "https://blogger.googleusercontent.com/img", "/img", 1),
					Space:  count,
				}

				table := h.DOM.Parent().Parent().Parent().Parent()

				if table.Is("tbody") {
					image.Description = table.Find(".tr-caption").Text()
				}

				post.Images[count] = image

			}
		})

		post.ID = e.Request.URL.Path[1 : len(e.Request.URL.Path)-5]
		e.ForEach("p", func(i int, h *colly.HTMLElement) {
			post.Body = append(post.Body, h.Text)
		})
	})

	c.Visit(url)

	return post
}

func fetchNewPosts(query string) []HomePost {

	var posts []HomePost

	c := colly.NewCollector()

	c.OnHTML(".blog-posts", func(h *colly.HTMLElement) {
		h.ForEach("article", func(i int, e *colly.HTMLElement) {

			id := e.ChildAttr(".entry-image-wrap", "href")
			image := e.ChildAttr(".entry-image", "data-image")

			post := HomePost{
				Heading: e.ChildText(".entry-title"),
				Image:   strings.Replace(image, "https://blogger.googleusercontent.com/img", "/img", 1),
				ID:      id[29 : len(id)-5],
			}

			posts = append(posts, post)
		})
	})

	if query == "" {
		c.Visit("https://www.der-postillon.com/search")
	} else {
		c.Visit("https://www.der-postillon.com/search?q=" + query)
	}

	return posts
}
