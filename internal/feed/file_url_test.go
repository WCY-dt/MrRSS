package feed

import (
	"strings"
	"testing"
)

func TestSanitizeFeedXML(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "Remove file:// atom:link",
			input: `<rss version="2.0">
				<channel>
					<title>Test Feed</title>
					<atom:link href="file://D:\tools\脚本语言\同步xml\rss_nyncb_tz_ai1.xml" rel="self" type="application/rss+xml" />
					<item>
						<title>Test Item</title>
					</item>
				</channel>
			</rss>`,
			shouldContain:    []string{"<title>Test Feed</title>", "<item>", "<title>Test Item</title>"},
			shouldNotContain: []string{`file://D:\tools\脚本语言\同步xml\rss_nyncb_tz_ai1.xml`},
		},
		{
			name: "Remove javascript: atom:link",
			input: `<rss version="2.0">
				<channel>
					<title>Test Feed</title>
					<atom:link href="javascript:void(0)" rel="self" type="application/rss+xml" />
					<item><title>Test</title></item>
				</channel>
			</rss>`,
			shouldContain:    []string{"<title>Test Feed</title>", "<item>"},
			shouldNotContain: []string{"javascript:void(0)"},
		},
		{
			name: "Remove data: atom:link",
			input: `<rss version="2.0">
				<channel>
					<title>Test Feed</title>
					<atom:link href="data:text/plain,test" rel="self" />
					<item><title>Test</title></item>
				</channel>
			</rss>`,
			shouldContain:    []string{"<title>Test Feed</title>"},
			shouldNotContain: []string{"data:text/plain"},
		},
		{
			name: "Remove ftp: atom:link",
			input: `<rss version="2.0">
				<channel>
					<title>Test Feed</title>
					<atom:link href="ftp://example.com/feed.xml" rel="self" />
					<item><title>Test</title></item>
				</channel>
			</rss>`,
			shouldContain:    []string{"<title>Test Feed</title>"},
			shouldNotContain: []string{"ftp://example.com"},
		},
		{
			name: "Keep http:// and https:// links",
			input: `<rss version="2.0">
				<channel>
					<title>Test Feed</title>
					<atom:link href="https://example.com/feed.xml" rel="self" type="application/rss+xml" />
					<link href="http://example.com/blog" />
					<item><title>Test</title></item>
				</channel>
			</rss>`,
			shouldContain: []string{
				"https://example.com/feed.xml",
				"http://example.com/blog",
			},
			shouldNotContain: []string{},
		},
		{
			name: "Remove standalone <link> with file://",
			input: `<rss version="2.0">
				<channel>
					<title>Test Feed</title>
					<link href="file:///local/feed.xml" />
					<item><title>Test</title></item>
				</channel>
			</rss>`,
			shouldContain:    []string{"<title>Test Feed</title>"},
			shouldNotContain: []string{"file:///local/feed.xml"},
		},
		{
			name: "Multiple file:// links",
			input: `<rss version="2.0">
				<channel>
					<title>Test Feed</title>
					<atom:link href="file://D:\feed1.xml" rel="self" />
					<atom:link href="file://D:\feed2.xml" rel="alternate" />
					<link href="file:///local/feed.xml" />
					<item><title>Test</title></item>
				</channel>
			</rss>`,
			shouldContain: []string{"<title>Test Feed</title>", "<item>"},
			shouldNotContain: []string{
				"file://D:\\feed1.xml",
				"file://D:\\feed2.xml",
				"file:///local/feed.xml",
			},
		},
		{
			name:             "Empty input",
			input:            "",
			shouldContain:    []string{},
			shouldNotContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFeedXML(tt.input)

			// Check shouldContain
			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("sanitizeFeedXML() result should contain %q\nInput:  %s\nResult: %s",
						expected, tt.input, result)
				}
			}

			// Check shouldNotContain
			for _, unexpected := range tt.shouldNotContain {
				if strings.Contains(result, unexpected) {
					t.Errorf("sanitizeFeedXML() result should NOT contain %q\nInput:  %s\nResult: %s",
						unexpected, tt.input, result)
				}
			}
		})
	}
}
