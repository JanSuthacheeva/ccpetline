package pet

import (
	"encoding/base64"
	"fmt"
	"strings"
)

const chunkSize = 4096

// KittyInlinePNG returns the escape sequence(s) to display a PNG inline
// using the Kitty graphics protocol. Uses q=2 to suppress terminal response.
func KittyInlinePNG(png []byte) string {
	b64 := base64.StdEncoding.EncodeToString(png)
	if len(b64) <= chunkSize {
		return fmt.Sprintf("\x1b_Ga=T,f=100,q=2;%s\x1b\\", b64)
	}
	var sb strings.Builder
	for i := 0; i < len(b64); i += chunkSize {
		end := i + chunkSize
		last := end >= len(b64)
		if last {
			end = len(b64)
		}
		m := 1
		if last {
			m = 0
		}
		if i == 0 {
			fmt.Fprintf(&sb, "\x1b_Ga=T,f=100,m=%d,q=2;%s\x1b\\", m, b64[i:end])
		} else {
			fmt.Fprintf(&sb, "\x1b_Gm=%d;%s\x1b\\", m, b64[i:end])
		}
	}
	return sb.String()
}
