package mdhtml

import (
	"github.com/vanng822/go-premailer/premailer"
	"gopkg.in/russross/blackfriday.v2"
)

// RenderMarkdown turn markdown to html
func RenderMarkdown(markdownContent string) (string, error) {
	htmlContent := githubCSS + string(blackfriday.Run([]byte(markdownContent)))
	prem := premailer.NewPremailerFromString(htmlContent, premailer.NewOptions())
	return prem.Transform()
}
