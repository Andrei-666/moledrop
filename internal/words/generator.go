package words

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func GenerateCode() string {
	//We use a random seed based on time to ensure we get different results each time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	adjective := adjectives[r.Intn(len(adjectives))]
	noun := nouns[r.Intn(len(nouns))]

	numStr := fmt.Sprintf("%d", r.Intn(10000))
	//we concatenate the code parts (two words and a number)
	parts := []string{adjective, noun, numStr}

	//we shuffle the parts to add more randomness to the code
	r.Shuffle(len(parts), func(i, j int) {
		parts[i], parts[j] = parts[j], parts[i]
	})

	return strings.Join(parts, "-")
}
