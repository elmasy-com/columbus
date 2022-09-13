package config

const (
	UserAgent = "Elmasy-Columbus/0.1-dev"
)

var (
	WorkingDir             string        // Directory to write
	Step                   int           // Number of logs queried at once
	BackgroundSaveInterval int    = 60   // Time to wait in seconds between two index save
	FetcherInterval        int    = 3600 // Time to wait in second between two successful fetch
	//CheckUnique          bool          // TODO: Check unique entry

)
