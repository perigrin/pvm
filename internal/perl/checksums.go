// ABOUTME: Checksum database for Perl source validation
// ABOUTME: Provides known good checksums for Perl releases to ensure integrity

package perl

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"tamarou.com/pvm/internal/xdg"
)

// ChecksumDatabase stores known checksums for Perl releases
type ChecksumDatabase struct {
	checksums map[string]string // version -> sha256 checksum
	mu        sync.RWMutex
}

// NewChecksumDatabase creates a new checksum database
func NewChecksumDatabase() (*ChecksumDatabase, error) {
	db := &ChecksumDatabase{
		checksums: make(map[string]string),
	}

	// Load embedded checksums
	if err := db.loadEmbeddedChecksums(); err != nil {
		return nil, err
	}

	// Load user checksums if available
	if err := db.loadUserChecksums(); err != nil {
		// Non-fatal, just log
		_ = err
	}

	return db, nil
}

// GetChecksum returns the checksum for a version
// First checks embedded checksums, then falls back to CPAN lookup
func (db *ChecksumDatabase) GetChecksum(version string) (string, error) {
	// First try embedded checksums
	db.mu.RLock()
	checksum, ok := db.checksums[version]
	db.mu.RUnlock()

	if ok {
		return checksum, nil
	}

	// Not found in embedded checksums, try CPAN fallback
	checksum, err := db.fetchChecksumFromCPAN(version)
	if err != nil {
		return "", fmt.Errorf("no checksum found for version %s: not in embedded checksums and CPAN lookup failed: %w", version, err)
	}

	// Cache the fetched checksum for future use
	db.mu.Lock()
	db.checksums[version] = checksum
	db.mu.Unlock()

	return checksum, nil
}

// AddChecksum adds a checksum for a version
func (db *ChecksumDatabase) AddChecksum(version, checksum string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.checksums[version] = checksum
}

// loadEmbeddedChecksums loads checksums from embedded data
func (db *ChecksumDatabase) loadEmbeddedChecksums() error {
	// Parse embedded checksum data
	scanner := bufio.NewScanner(strings.NewReader(embeddedChecksums))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		checksum := parts[0]
		filename := parts[1]

		// Extract version from filename
		version := extractVersionFromFilename(filename)
		if version != "" {
			db.checksums[version] = checksum
		}
	}

	return scanner.Err()
}

// loadUserChecksums loads user-provided checksums from XDG_CONFIG_HOME/pvm/checksums.txt
func (db *ChecksumDatabase) loadUserChecksums() error {
	// Get XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return fmt.Errorf("failed to get XDG directories: %w", err)
	}

	// Build path to user checksums file
	checksumsFile := filepath.Join(dirs.ConfigDir, "checksums.txt")

	// Check if file exists
	if _, err := os.Stat(checksumsFile); os.IsNotExist(err) {
		// File doesn't exist, which is fine - no user checksums
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat checksums file %s: %w", checksumsFile, err)
	}

	// Open and read the file
	file, err := os.Open(checksumsFile)
	if err != nil {
		return fmt.Errorf("failed to open checksums file %s: %w", checksumsFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse line format: version algorithm checksum
		parts := strings.Fields(line)
		if len(parts) != 3 {
			return fmt.Errorf("invalid checksum format at line %d in %s: expected 'version algorithm checksum', got: %s",
				lineNumber, checksumsFile, line)
		}

		version := parts[0]
		algorithm := strings.ToLower(parts[1])
		checksum := parts[2]

		// For now, only support sha256 (can be extended later)
		if algorithm != "sha256" {
			return fmt.Errorf("unsupported checksum algorithm '%s' at line %d in %s (currently only sha256 is supported)",
				algorithm, lineNumber, checksumsFile)
		}

		// Validate checksum format (sha256 should be 64 hex characters)
		if len(checksum) != 64 {
			return fmt.Errorf("invalid sha256 checksum length at line %d in %s: expected 64 characters, got %d",
				lineNumber, checksumsFile, len(checksum))
		}

		// Store in database (user checksums override embedded ones)
		db.mu.Lock()
		db.checksums[version] = checksum
		db.mu.Unlock()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading checksums file %s: %w", checksumsFile, err)
	}

	return nil
}

// extractVersionFromFilename extracts version from perl-X.Y.Z.tar.gz filename
func extractVersionFromFilename(filename string) string {
	// Remove extension
	name := filename
	for _, ext := range []string{".tar.gz", ".tar.xz", ".tgz"} {
		if strings.HasSuffix(name, ext) {
			name = name[:len(name)-len(ext)]
			break
		}
	}

	// Extract version
	if strings.HasPrefix(name, "perl-") {
		return name[5:]
	}

	return ""
}

// embeddedChecksums contains known checksums for Perl releases
// Format: sha256sum  filename
// Source: https://www.cpan.org/src/5.0/CHECKSUMS
const embeddedChecksums = `
# Perl 5.42.x series
73cf6cc1ea2b2b1c110a18c14bbbc73a362073003893ffcedc26d22ebdbdd0c3  perl-5.42.0.tar.xz

# Perl 5.40.x series
0551c717458e703ef7972307ab19385edfa231198d88998df74e12226abf563b  perl-5.40.2.tar.xz
4ab18a5d642c1085fd41e52de89eec4cbf4a3a47cd2ac6178608e5f0ef1fa76f  perl-5.40.0.tar.xz
82b2a4c29cb6c4a2910941d89d2308737c23a8b26e7c2b87238496dd2a2e7b69  perl-5.40.1.tar.xz

# Perl 5.38.x series
eca551caec3bc549a4e590c0015003790bdd1a604ffe19cc78ee631d51f7072e  perl-5.38.2.tar.xz
0684385a2952497db4df9e32b429e01527ca69c31a3b327a88fae90fb1511721  perl-5.38.1.tar.xz
9c9a8a1fb7b6b330a13e9b7c4829e9891b832779c89698cb5c307c19b0ce9567  perl-5.38.0.tar.xz

# Perl 5.36.x series
eb8a026f5d68e9b982e08e8b8b11fc92897337d8c50b4f0017b3d0076b99d1d5  perl-5.36.3.tar.xz
a57ba0e7414bb17ae73eed7a889e5c1e632a1e1cb80ac29d30d9e1d7a96fe673  perl-5.36.2.tar.xz
6166488d7dff7a0ff0c9cb9e1f8bf1e2e5da5f8c0c04d28b4e76bd5243b0e0e8  perl-5.36.1.tar.xz
d3e5c07fa43eef52178e9c9113ee802e0a6e82c96b594cf6fba87dd6a7dd5a09  perl-5.36.0.tar.xz

# Perl 5.34.x series
65d09f3cb8e7c188cbe99c1b850dcb7b98bcdff2aca03acb3f69fa660b07f45d  perl-5.34.3.tar.xz
4ac7dc2d37e171ccd21a5dc6e373f7b7d8444641d975f9322a7dd5ad2c8ddcb1  perl-5.34.2.tar.xz
fdb22d3320f480a1dc233db67e2f3de97248b8dc39c4d8242d577af92218e525  perl-5.34.1.tar.xz
7ba2a9551c3e49b873ce427f80393eb0a4669d5ce8d8bb903c5a22c6a23e3405  perl-5.34.0.tar.xz

# Perl 5.32.x series
6f436b447cf56d22464f980fac1916e707a040e96d52172984c5d184c09b859b  perl-5.32.1.tar.xz
efeb1ce1f10824190ad1cadbcccf6fdb8a5d37007d0100d2d9ae5f2b5900c0b4  perl-5.32.0.tar.xz

# Perl 5.30.x series
bf3d25571ff1ee94186177c2cdef87867fd6a14aa5a84f0b1fb7bf798f42f964  perl-5.30.3.tar.xz
6967595f2e3f3a94544c35152f9a25e0cb8ea24ae45f4bf1882f2e33f4a400f4  perl-5.30.2.tar.xz
7b3ce23ec8b4f45e7bb7642842e96a1dae5a2d7d80893d5b065b96e2cd76e858  perl-5.30.1.tar.xz
ac501cad4af904d33370a9ea39dbb7a8ad4cb19bc7bc8a9c17d8dc3e81ef6306  perl-5.30.0.tar.xz

# Perl 5.28.x series (older format)
5ce0520cdfdcae39ca48659feb51f04de72cd43c5329e74eb42322149af95010  perl-5.28.3.tar.gz
9165a1b290fcd59f17170dd9dc248133b82fa3e30563d040d2109dc846ae80a8  perl-5.28.2.tar.gz
3ebf85fe65df2ee165b22596540b7d5d42f84d4b72d84834f74e2e0b8956c347  perl-5.28.1.tar.gz
7e929f64d4cb0e9d1159d4a59fc89394e27fa1f7004d0836ca0d514685406ea8  perl-5.28.0.tar.gz

# Perl 5.26.x series
203afca8995ca426db0af48b78eb606b5d24011a307b60bc77aaa296c9f14a54  perl-5.26.3.tar.gz
22b00824d9f7762531ceb73bc4c7c06e6f600de6c364e3cbfa17071b2eb5d3c9  perl-5.26.2.tar.gz
fe8208133e73e47afc3251c08d2c21c5a60160165a8ab8b669c43a420e4ec680  perl-5.26.1.tar.gz
9bf2e3d0d72aad77865c3bdbc20d3b576d769c5c255c4ceb30fdb9335266bf55  perl-5.26.0.tar.gz
`

// checksumsFS would be used for external checksum files
// var checksumsFS embed.FS

// fetchChecksumFromCPAN fetches the checksum for a specific version from CPAN
func (db *ChecksumDatabase) fetchChecksumFromCPAN(version string) (string, error) {
	// CPAN CHECKSUMS URL
	checksumsURL := "https://www.cpan.org/src/5.0/CHECKSUMS"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch the CHECKSUMS file
	resp, err := client.Get(checksumsURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch CHECKSUMS from CPAN: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CPAN CHECKSUMS request failed with status %d", resp.StatusCode)
	}

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read CHECKSUMS response: %w", err)
	}

	// Parse the CHECKSUMS file to find our version
	checksum, err := db.parseChecksumFromCPANData(string(body), version)
	if err != nil {
		return "", fmt.Errorf("failed to parse checksum for version %s: %w", version, err)
	}

	return checksum, nil
}

// parseChecksumFromCPANData parses the CPAN CHECKSUMS format to extract checksum for a version
func (db *ChecksumDatabase) parseChecksumFromCPANData(data, version string) (string, error) {
	// The CPAN CHECKSUMS file has entries like:
	// 'perl-5.42.0.tar.xz' => {
	//   'sha256' => 'abc123...',
	//   'size' => 12345,
	// },

	// Create regex to match the perl version entry
	filename := fmt.Sprintf("perl-%s.tar.xz", version)

	// Look for the sha256 entry for this file
	// Pattern matches: 'perl-X.Y.Z.tar.xz' => { ... 'sha256' => 'checksum', ... }
	pattern := fmt.Sprintf(`'%s'\s*=>\s*\{[^}]*'sha256'\s*=>\s*'([a-f0-9]{64})'`, regexp.QuoteMeta(filename))
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(data)
	if len(matches) < 2 {
		// Try .tar.gz as fallback
		filenameGz := fmt.Sprintf("perl-%s.tar.gz", version)
		patternGz := fmt.Sprintf(`'%s'\s*=>\s*\{[^}]*'sha256'\s*=>\s*'([a-f0-9]{64})'`, regexp.QuoteMeta(filenameGz))
		reGz := regexp.MustCompile(patternGz)

		matches = reGz.FindStringSubmatch(data)
		if len(matches) < 2 {
			return "", fmt.Errorf("checksum not found in CPAN CHECKSUMS for perl-%s", version)
		}
	}

	return matches[1], nil
}

// UpdateChecksums updates the checksum database from CPAN
func (db *ChecksumDatabase) UpdateChecksums() error {
	// This would fetch the latest CHECKSUMS file from CPAN
	// and update the local database
	// For now, we rely on embedded checksums + fallback
	return fmt.Errorf("bulk checksum updates not yet implemented")
}
