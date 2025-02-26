package parse_jwt

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sikalabs/slr/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "parse-jwt",
	Short: "Parse JWT from stdin into JSON list of 3 objects (Header, Payload, Signature)",
	Run: func(cmd *cobra.Command, args []string) {
		jwtToken := readFromPipe()
		parseJWT(jwtToken)
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
}

func parseJWT(jwtToken string) {
	// Parse JWT
	token, _ := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) { return nil, nil })

	// Marshal header and claims to JSON
	headerJSON, _ := json.Marshal(token.Header)
	claimsJSON, _ := json.Marshal(token.Claims)

	// Prepare the result as a slice of interfaces
	result := []interface{}{
		decodeJSON(headerJSON),
		decodeJSON(claimsJSON),
		token.Signature,
	}

	// Print the result as a JSON array
	outputJSON, err := json.Marshal(result)
	if err != nil {
		log.Fatal("Error marshalling result to JSON: ", err)
	}
	fmt.Println(string(outputJSON))
}

// decodeJSON unmarshals JSON bytes into an empty interface
func decodeJSON(data []byte) interface{} {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		log.Fatal("Error unmarshalling JSON: ", err)
	}
	return obj
}

func readFromPipe() string {
	var jwtToken string

	// Ensure input is from a pipe
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeNamedPipe == 0 {
		log.Fatalln("No input from pipe.")
	}

	// Read the input from stdin (pipe)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		jwtToken = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading from stdin: ", err)
	}

	return jwtToken
}
