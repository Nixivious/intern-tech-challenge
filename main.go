package main

import (
	"context"
	"fmt"
	"os"
	"bufio"
	"strings"
	"sort"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/github"
)

// LatestVersions returns a sorted slice with the highest version as its first element and the highest version of the smaller minor versions in a descending order
func LatestVersions(releases []*semver.Version, minVersion *semver.Version) []*semver.Version {
	var versionSlice []*semver.Version

	// Sort release array
	sort.Sort(semver.Versions(releases))

	// Compare each release, add to versionSlice if of latest release in the same minor version (major.minor.x)
	var major, minor int64 // Keeps track of the latest version of major or minor
	var prev *semver.Version
	for _, release := range releases {
		// Consider versions above specified minVersion.
		if !release.LessThan(*minVersion) {
			// Initialize here instead since first row of releases may be larger than minVersion
			if prev == nil {
				major, minor = release.Major, release.Minor
				prev = release
			}

			// We take advantage of the sorted releases array, the current patch must be greater than prev
			if major < release.Major || minor < release.Minor { // if either major or minor increments, we observe an increment in version, and so the previous patch observed is the greatest of the previous version
				major, minor = release.Major, release.Minor
				versionSlice = append(versionSlice, prev)
			}

			prev = release
		}
	}

	if prev != nil {
		versionSlice = append(versionSlice, prev)
	}

	// Sort release array in reverse order
	sort.Sort(sort.Reverse(semver.Versions(versionSlice)))

	return versionSlice
}

// Here we implement the basics of communicating with github through the library as well as printing the version
// You will need to implement LatestVersions function as well as make this application support the file format outlined in the README
// Please use the format defined by the fmt.Printf line at the bottom, as we will define a passing coding challenge as one that outputs
// the correct information, including this line
func main() {

	// Check command line arguments
	args := os.Args
	if len(args) != 2 {
		fmt.Println("Input file required, please specify path! (expected: arg1:file-path)")
		return
	}

	// Open file for parsing
	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Repo information storage
	var repo_Owners 		[]string
	var repo_Names  		[]string
	var repo_mVersions 		[]*semver.Version

	// Iterate through file, extract variables
	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		line 		:= scanner.Text()

		// Split each csv, assumes (repository,min_version)
		lineParts 	:= strings.Split(line,",")
		if len(lineParts) != 2 { // Ignores incorrectly formatted line
			continue
		}

		// Split repository into its owner and name parts
		repoParts	:= strings.Split(lineParts[0],"/")
		if len(repoParts) != 2 { // Ignores incorrectly formatted repo, filters out the first line (repository,min_version)
			continue
		}
		name 		:= repoParts[0]
		owner 		:= repoParts[1]

		// Create semver object for min version from file
		version, err	:= semver.NewVersion(lineParts[1])
		if err != nil { // Ignore semver err
			continue
		}

		// Append results into storage
		repo_Owners 	= append(repo_Owners, name)
		repo_Names 		= append(repo_Names, owner)
		repo_mVersions 	= append(repo_mVersions, version)
	}

	// Github
	client := github.NewClient(nil)
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 10}

	// Loop through each repo
	for i := 0; i < len(repo_Owners); i++ {
		// Get release list from GitHub
		releases, _, err := client.Repositories.ListReleases(ctx, repo_Owners[i], repo_Names[i], opt)
		if err != nil { // Handle err, skip repo so not to interrupt loop
			fmt.Printf("Error in retrieving list releases! \n %s \n", err)
			continue
		}

		minVersion 	:= repo_mVersions[i]
		allReleases := make([]*semver.Version, len(releases))

		// Create semver objects for each release
		for i, release := range releases {
			versionString := *release.TagName
			if versionString[0] == 'v' {
				versionString = versionString[1:]
			}
			allReleases[i] = semver.New(versionString)
		}
		// Get latest patch for each release
		versionSlice := LatestVersions(allReleases, minVersion)

		// Output Result
		fmt.Printf("latest versions of %s/%s: %s\n", repo_Owners[i], repo_Names[i], versionSlice)
	}
}