package conda

import (
	"encoding/json"
	"io"
	"strings"
)

// PackageExistsInChannelData streams channeldata.json and checks if package exists
// Returns the package info if found, nil if not found
func PackageExistsInChannelData(r io.Reader, pkgName string) (*PackageInfo, bool) {
	decoder := json.NewDecoder(r)
	
	// Look for "packages" key
	for {
		tok, err := decoder.Token()
		if err != nil {
			return nil, false
		}
		
		if str, ok := tok.(string); ok && str == "packages" {
			// Enter the packages object
			tok, err = decoder.Token()
			if err != nil || tok != json.Delim('{') {
				return nil, false
			}
			
			// Stream through package names
			for decoder.More() {
				name, err := decoder.Token()
				if err != nil {
					return nil, false
				}
				
				nameStr, ok := name.(string)
				if !ok {
					return nil, false
				}
				
				// Check if this is our package
				if strings.EqualFold(nameStr, pkgName) {
					// Found it! Parse the package info
					var pkgInfo PackageInfo
					if err := decoder.Decode(&pkgInfo); err != nil {
						return nil, false
					}
					return &pkgInfo, true
				}
				
				// Skip this package's value
				if err := skipValue(decoder); err != nil {
					return nil, false
				}
			}
			
			// Package not found
			return nil, false
		}
	}
}

// PackageExistsInRepoData streams repodata.json and checks if package exists
// Returns the package info if found, nil if not found
func PackageExistsInRepoData(r io.Reader, pkgName string) (*Package, bool) {
	decoder := json.NewDecoder(r)
	pkgName = strings.ToLower(pkgName)
	
	for {
		tok, err := decoder.Token()
		if err != nil {
			return nil, false
		}
		
		if str, ok := tok.(string); ok {
			// Check for "packages" or "packages.conda"
			if str == "packages" || str == "packages.conda" {
				// Enter the object
				tok, err = decoder.Token()
				if err != nil || tok != json.Delim('{') {
					return nil, false
				}
				
				// Stream through packages
				for decoder.More() {
					// Skip filename key
					_, err := decoder.Token()
					if err != nil {
						return nil, false
					}
					
					// Parse package
					var pkg Package
					if err := decoder.Decode(&pkg); err != nil {
						continue
					}
					
					// Check if name matches
					if strings.ToLower(pkg.Name) == pkgName {
						return &pkg, true
					}
				}
			}
		}
	}
}

// skipValue skips over a JSON value (object, array, or primitive)
func skipValue(decoder *json.Decoder) error {
	tok, err := decoder.Token()
	if err != nil {
		return err
	}
	
	switch tok {
	case json.Delim('{'):
		// Skip object
		for decoder.More() {
			// Skip key
			if _, err := decoder.Token(); err != nil {
				return err
			}
			// Skip value
			if err := skipValue(decoder); err != nil {
				return err
			}
		}
		// Skip closing }
		_, err = decoder.Token()
		return err
		
	case json.Delim('['):
		// Skip array
		for decoder.More() {
			if err := skipValue(decoder); err != nil {
				return err
			}
		}
		// Skip closing ]
		_, err = decoder.Token()
		return err
		
	default:
		// Primitive value, already consumed
		return nil
	}
}

// CollectPackageNamesFromRepoData streams repodata.json and collects unique package names
func CollectPackageNamesFromRepoData(r io.Reader) (map[string]bool, error) {
	decoder := json.NewDecoder(r)
	names := make(map[string]bool)
	
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			return names, nil
		}
		if err != nil {
			return nil, err
		}
		
		if str, ok := tok.(string); ok {
			if str == "packages" || str == "packages.conda" {
				tok, err = decoder.Token()
				if err != nil || tok != json.Delim('{') {
					return names, nil
				}
				
				for decoder.More() {
					// Skip filename key
					_, err := decoder.Token()
					if err != nil {
						return names, nil
					}
					
					// Parse just the name field
					var pkg struct {
						Name string `json:"name"`
					}
					if err := decoder.Decode(&pkg); err != nil {
						continue
					}
					if pkg.Name != "" {
						names[strings.ToLower(pkg.Name)] = true
					}
				}
			}
		}
	}
}

// CollectPackageNamesFromChannelData streams channeldata.json and collects package names
func CollectPackageNamesFromChannelData(r io.Reader) (map[string]bool, error) {
	decoder := json.NewDecoder(r)
	names := make(map[string]bool)
	
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			return names, nil
		}
		if err != nil {
			return nil, err
		}
		
		if str, ok := tok.(string); ok && str == "packages" {
			tok, err = decoder.Token()
			if err != nil || tok != json.Delim('{') {
				return names, nil
			}
			
			for decoder.More() {
				name, err := decoder.Token()
				if err != nil {
					return names, nil
				}
				
				if nameStr, ok := name.(string); ok {
					names[strings.ToLower(nameStr)] = true
				}
				
				// Skip package value
				if err := skipValue(decoder); err != nil {
					return names, nil
				}
			}
		}
	}
}
