package presets

type Preset struct {
	Name        string
	URL         string
	Description string
	Category    string
}

var curatedPresets = []Preset{
	{
		Name:        "JSONPlaceholder Todo",
		URL:         "https://jsonplaceholder.typicode.com/todos/1",
		Description: "Simple mock task object",
		Category:    "Mock",
	},
	{
		Name:        "JSONPlaceholder Post",
		URL:         "https://jsonplaceholder.typicode.com/posts/1",
		Description: "Mock blog post with title and body",
		Category:    "Mock",
	},
	{
		Name:        "GitHub API (Octocat)",
		URL:         "https://api.github.com/users/octocat",
		Description: "GitHub user profile metadata",
		Category:    "Dev",
	},
	{
		Name:        "GitHub API (Golang Repo)",
		URL:         "https://api.github.com/repos/golang/go",
		Description: "Repository statistics, stars & forks",
		Category:    "Dev",
	},
	{
		Name:        "HTTPBin IP",
		URL:         "https://httpbin.org/ip",
		Description: "Inspect client public IP",
		Category:    "Utility",
	},
	{
		Name:        "HTTPBin Headers",
		URL:         "https://httpbin.org/headers",
		Description: "Echo back sent request headers",
		Category:    "Utility",
	},
	{
		Name:        "PokeAPI (Pikachu)",
		URL:         "https://pokeapi.co/api/v2/pokemon/pikachu",
		Description: "Pokemon stats, abilities and sprites",
		Category:    "Fun",
	},
	{
		Name:        "Rick & Morty (Morty)",
		URL:         "https://rickandmortyapi.com/api/character/2",
		Description: "Character specs, origin and episodes",
		Category:    "Fun",
	},
	{
		Name:        "Cat Fact",
		URL:         "https://catfact.ninja/fact",
		Description: "Random cat trivia",
		Category:    "Fun",
	},
	{
		Name:        "DummyJSON Product",
		URL:         "https://dummyjson.com/products/1",
		Description: "E-commerce product catalog record",
		Category:    "Mock",
	},
}

func GetPresets() []Preset {
	return curatedPresets
}
