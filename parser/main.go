package main
import(
	"github.com/russross/blackfriday"	// russross parser
	"strings"							// string replace
	"regexp"							// regex
	"fmt"			// testing1
	"os"			// testing2
)

func parse(inp string) string {
	inp = strings.Replace(inp,`- [ ]`,`- <input type="checkbox">`,-1)			// Set Unchecked Checkbox
	inp = strings.Replace(inp,`- [x]`,`- <input type="checkbox" checked>`,-1)	// Set Checked Checkbox
	data := []byte(inp)															// Convert to Byte
	regex , _ := regexp.Compile("[sS][cC][rR][iI][pP][tT]")						// Escape Script Tag
	data = regex.ReplaceAll(data,[]byte("&#115;&#99;&#114;&#105;&#112;&#116;")) 
	data = blackfriday.MarkdownCommon(data)										// Get Common Markdown from russross's parser
	return string(data)
}

func main(){
	file, err := os.Open("test.txt")	// Open Test File
	if err != nil { panic(-1); }		// If it doesn't open then panic.
	buffer := make([]byte,2000)			// Make buffer to hold output.
	file.Read(buffer)					// Read file contents in.
	fmt.Println(parse(string(buffer)))  // Log out the parsed file.
}