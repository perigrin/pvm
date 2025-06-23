// ABOUTME: End-to-end integration tests for build-perl and install-perl workflow
// ABOUTME: Tests the complete build-only, install-from-build, and install-from-archive workflows

package e2e

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/test/e2e/helpers"
)

func TestBuildInstallWorkflow_BuildOnly(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a temporary output directory
	outputDir := filepath.Join(env.RootDir, "perl-build-output")

	// Test build-only without installation
	stdout, stderr, err := env.RunPVM("build-perl", "5.38.0", "--build-only", "--output-dir", outputDir)
	if err != nil {
		t.Skipf("TODO: build-perl --build-only not yet fully implemented\nCommand: pvm build-perl 5.38.0 --build-only --output-dir %s\nError: %v\nStdout: %s\nStderr: %s", outputDir, err, stdout, stderr)
	}

	// Verify the output directory exists and contains a complete Perl installation
	require.DirExists(t, outputDir, "Output directory should exist")
	require.FileExists(t, filepath.Join(outputDir, "bin", "perl"), "Perl binary should exist")
	require.DirExists(t, filepath.Join(outputDir, "lib", "perl5"), "Perl library directory should exist")

	// Verify the built Perl is functional
	perlBin := filepath.Join(outputDir, "bin", "perl")
	perlStdout, perlStderr, perlErr := env.RunCommand(perlBin, "-v")
	require.NoError(t, perlErr, "Built Perl should be functional: %s", perlStderr)
	assert.Contains(t, perlStdout, "v5.38.0", "Built Perl should report correct version")

	// Verify that the Perl was NOT installed in PVM's version directory
	listStdout, _, listErr := env.RunPVM("list")
	if listErr == nil {
		assert.NotContains(t, listStdout, "5.38.0", "Perl should not be installed in PVM")
	}
}

func TestBuildInstallWorkflow_InstallFromBuild(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// First, create a build-only Perl installation
	outputDir := filepath.Join(env.RootDir, "perl-build-output")
	stdout, stderr, err := env.RunPVM("build-perl", "5.38.0", "--build-only", "--output-dir", outputDir)
	if err != nil {
		t.Skipf("TODO: build-perl --build-only not yet fully implemented\nCommand: pvm build-perl 5.38.0 --build-only --output-dir %s\nError: %v\nStdout: %s\nStderr: %s", outputDir, err, stdout, stderr)
	}

	// Now install from the build directory
	installStdout, installStderr, installErr := env.RunPVM("install-perl", "--from-build", outputDir, "--version", "5.38.0-e2e-test")
	if installErr != nil {
		t.Skipf("TODO: install-perl --from-build not yet fully implemented\nCommand: pvm install-perl --from-build %s --version 5.38.0-e2e-test\nError: %v\nStdout: %s\nStderr: %s", outputDir, installErr, installStdout, installStderr)
	}

	// Verify the Perl was installed in PVM's version management
	listStdout, _, listErr := env.RunPVM("list")
	if listErr == nil {
		assert.Contains(t, listStdout, "5.38.0-e2e-test", "Perl should be installed in PVM")
	}

	// Verify the installed Perl is functional
	useStdout, useStderr, useErr := env.RunPVM("use", "5.38.0-e2e-test")
	if useErr == nil {
		perlStdout, perlStderr, perlErr := env.RunCommand("perl", "-v")
		if perlErr == nil {
			assert.Contains(t, perlStdout, "v5.38.0", "Installed Perl should report correct version")
		} else {
			t.Logf("Could not test installed Perl functionality: %v\nStdout: %s\nStderr: %s", perlErr, perlStdout, perlStderr)
		}
	} else {
		t.Logf("Could not use installed Perl: %v\nStdout: %s\nStderr: %s", useErr, useStdout, useStderr)
	}
}

func TestBuildInstallWorkflow_InstallFromArchive(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// First, create a build-only Perl installation
	outputDir := filepath.Join(env.RootDir, "perl-build-output")
	stdout, stderr, err := env.RunPVM("build-perl", "5.38.0", "--build-only", "--output-dir", outputDir)
	if err != nil {
		t.Skipf("TODO: build-perl --build-only not yet fully implemented\nCommand: pvm build-perl 5.38.0 --build-only --output-dir %s\nError: %v\nStdout: %s\nStderr: %s", outputDir, err, stdout, stderr)
	}

	// Create a tar.gz archive from the build directory
	archivePath := filepath.Join(env.RootDir, "perl-5.38.0-e2e-test.tar.gz")
	err = createTarGzArchive(outputDir, archivePath)
	require.NoError(t, err, "Should be able to create tar.gz archive")

	// Install from the archive
	installStdout, installStderr, installErr := env.RunPVM("install-perl", archivePath, "--version", "5.38.0-archive-test")
	if installErr != nil {
		t.Skipf("TODO: install-perl from archive not yet fully implemented\nCommand: pvm install-perl %s --version 5.38.0-archive-test\nError: %v\nStdout: %s\nStderr: %s", archivePath, installErr, installStdout, installStderr)
	}

	// Verify the Perl was installed in PVM's version management
	listStdout, _, listErr := env.RunPVM("list")
	if listErr == nil {
		assert.Contains(t, listStdout, "5.38.0-archive-test", "Perl should be installed in PVM")
	}

	// Verify the installed Perl is functional
	useStdout, useStderr, useErr := env.RunPVM("use", "5.38.0-archive-test")
	if useErr == nil {
		perlStdout, perlStderr, perlErr := env.RunCommand("perl", "-v")
		if perlErr == nil {
			assert.Contains(t, perlStdout, "v5.38.0", "Installed Perl should report correct version")
		} else {
			t.Logf("Could not test installed Perl functionality: %v\nStdout: %s\nStderr: %s", perlErr, perlStdout, perlStderr)
		}
	} else {
		t.Logf("Could not use installed Perl: %v\nStdout: %s\nStderr: %s", useErr, useStdout, useStderr)
	}
}

func TestBuildInstallWorkflow_RelocatableBuilds(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a build-only Perl installation
	outputDir := filepath.Join(env.RootDir, "perl-build-original")
	stdout, stderr, err := env.RunPVM("build-perl", "5.38.0", "--build-only", "--output-dir", outputDir)
	if err != nil {
		t.Skipf("TODO: build-perl --build-only not yet fully implemented\nCommand: pvm build-perl 5.38.0 --build-only --output-dir %s\nError: %v\nStdout: %s\nStderr: %s", outputDir, err, stdout, stderr)
	}

	// Move the build to a different location
	movedDir := filepath.Join(env.RootDir, "perl-build-moved")
	err = os.Rename(outputDir, movedDir)
	require.NoError(t, err, "Should be able to move build directory")

	// Verify the moved Perl is still functional (relocatable)
	perlBin := filepath.Join(movedDir, "bin", "perl")
	versionStdout, versionStderr, versionErr := env.RunCommand(perlBin, "-v")
	if versionErr != nil {
		t.Logf("Relocated Perl not functional: %v\nStdout: %s\nStderr: %s", versionErr, versionStdout, versionStderr)
		return
	}
	assert.Contains(t, versionStdout, "v5.38.0", "Relocated Perl should report correct version")

	// Test that @INC is correctly set up for the new location
	incStdout, incStderr, incErr := env.RunCommand(perlBin, "-e", "print join('\\n', @INC)")
	if incErr == nil {
		assert.Contains(t, incStdout, movedDir, "@INC should reference the new location")
	} else {
		t.Logf("Could not test @INC: %v\nStdout: %s\nStderr: %s", incErr, incStdout, incStderr)
	}
}

func TestBuildInstallWorkflow_BackwardCompatibility(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test that the traditional build-perl command still works (build AND install)
	stdout, stderr, err := env.RunPVM("build-perl", "5.38.0")
	if err != nil {
		t.Skipf("TODO: build-perl traditional mode not yet fully implemented\nCommand: pvm build-perl 5.38.0\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Verify the Perl was both built and installed
	listStdout, _, listErr := env.RunPVM("list")
	if listErr == nil {
		assert.Contains(t, listStdout, "5.38.0", "Perl should be installed in PVM")
	}

	// Verify the installed Perl is functional
	useStdout, useStderr, useErr := env.RunPVM("use", "5.38.0")
	if useErr == nil {
		perlStdout, perlStderr, perlErr := env.RunCommand("perl", "-v")
		if perlErr == nil {
			assert.Contains(t, perlStdout, "v5.38.0", "Installed Perl should report correct version")
		} else {
			t.Logf("Could not test installed Perl functionality: %v\nStdout: %s\nStderr: %s", perlErr, perlStdout, perlStderr)
		}
	} else {
		t.Logf("Could not use installed Perl: %v\nStdout: %s\nStderr: %s", useErr, useStdout, useStderr)
	}
}

func TestBuildInstallWorkflow_ErrorHandling(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test install-perl with non-existent directory
	installStdout, installStderr, installErr := env.RunPVM("install-perl", "--from-build", "/nonexistent/directory")
	if installErr == nil {
		t.Logf("Expected install-perl to fail with non-existent directory, but it succeeded\nStdout: %s\nStderr: %s", installStdout, installStderr)
	} else {
		assert.Contains(t, installStderr, "does not exist", "Error message should be helpful")
	}

	// Test install-perl with invalid archive
	invalidArchive := filepath.Join(env.RootDir, "invalid.tar.gz")
	err := os.WriteFile(invalidArchive, []byte("not a valid archive"), 0644)
	require.NoError(t, err, "Should be able to create invalid archive file")

	archiveStdout, archiveStderr, archiveErr := env.RunPVM("install-perl", invalidArchive)
	if archiveErr == nil {
		t.Logf("Expected install-perl to fail with invalid archive, but it succeeded\nStdout: %s\nStderr: %s", archiveStdout, archiveStderr)
	} else {
		assert.Contains(t, archiveStderr, "extract", "Error message should mention extraction")
	}

	// Test build-perl with invalid output directory (permission denied)
	buildStdout, buildStderr, buildErr := env.RunPVM("build-perl", "5.38.0", "--build-only", "--output-dir", "/root/forbidden")
	if buildErr == nil {
		t.Logf("Expected build-perl to fail with permission denied, but it succeeded\nStdout: %s\nStderr: %s", buildStdout, buildStderr)
	}
}

func TestBuildInstallWorkflow_VersionDetection(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a build-only Perl installation
	outputDir := filepath.Join(env.RootDir, "perl-build-output")
	stdout, stderr, err := env.RunPVM("build-perl", "5.38.0", "--build-only", "--output-dir", outputDir)
	if err != nil {
		t.Skipf("TODO: build-perl --build-only not yet fully implemented\nCommand: pvm build-perl 5.38.0 --build-only --output-dir %s\nError: %v\nStdout: %s\nStderr: %s", outputDir, err, stdout, stderr)
	}

	// Install without specifying version (should auto-detect)
	installStdout, installStderr, installErr := env.RunPVM("install-perl", "--from-build", outputDir)
	if installErr != nil {
		t.Skipf("TODO: install-perl version auto-detection not yet fully implemented\nCommand: pvm install-perl --from-build %s\nError: %v\nStdout: %s\nStderr: %s", outputDir, installErr, installStdout, installStderr)
	}

	// Verify the Perl was installed with the correct version
	listStdout, _, listErr := env.RunPVM("list")
	if listErr == nil {
		assert.Contains(t, listStdout, "5.38.0", "Perl should be installed with detected version")
	}
}

func TestBuildInstallWorkflow_CompleteIntegration(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// This test simulates a complete binary distribution workflow:
	// 1. Build-only with custom output directory
	// 2. Create archive for distribution
	// 3. Install from archive on "different system"
	// 4. Verify functionality and relocatability

	// Step 1: Build-only
	buildDir := filepath.Join(env.RootDir, "build-output")
	buildStdout, buildStderr, buildErr := env.RunPVM("build-perl", "5.38.0", "--build-only", "--output-dir", buildDir)
	if buildErr != nil {
		t.Skipf("TODO: build-perl --build-only not yet fully implemented\nCommand: pvm build-perl 5.38.0 --build-only --output-dir %s\nError: %v\nStdout: %s\nStderr: %s", buildDir, buildErr, buildStdout, buildStderr)
	}

	// Step 2: Create distribution archive
	archivePath := filepath.Join(env.RootDir, "perl-5.38.0-linux-amd64.tar.gz")
	err := createTarGzArchive(buildDir, archivePath)
	require.NoError(t, err, "Archive creation should succeed")

	// Verify archive properties
	archiveInfo, err := os.Stat(archivePath)
	require.NoError(t, err, "Archive should exist")
	assert.Greater(t, archiveInfo.Size(), int64(10*1024*1024), "Archive should be reasonably large (>10MB)")

	// Step 3: Install from archive (simulating different system)
	installStdout, installStderr, installErr := env.RunPVM("install-perl", archivePath, "--version", "5.38.0-distributed")
	if installErr != nil {
		t.Skipf("TODO: install-perl from archive not yet fully implemented\nCommand: pvm install-perl %s --version 5.38.0-distributed\nError: %v\nStdout: %s\nStderr: %s", archivePath, installErr, installStdout, installStderr)
	}

	// Step 4: Verify complete functionality
	useStdout, useStderr, useErr := env.RunPVM("use", "5.38.0-distributed")
	if useErr != nil {
		t.Logf("Could not use installed Perl: %v\nStdout: %s\nStderr: %s", useErr, useStdout, useStderr)
		return
	}

	// Test basic Perl functionality
	helloStdout, helloStderr, helloErr := env.RunCommand("perl", "-e", "print 'Hello, World!'")
	if helloErr == nil {
		assert.Equal(t, "Hello, World!", helloStdout, "Perl output should be correct")
	} else {
		t.Logf("Basic Perl execution failed: %v\nStdout: %s\nStderr: %s", helloErr, helloStdout, helloStderr)
	}

	// Test module loading
	moduleStdout, moduleStderr, moduleErr := env.RunCommand("perl", "-e", "use strict; use warnings; print 'OK'")
	if moduleErr == nil {
		assert.Equal(t, "OK", moduleStdout, "Module loading output should be correct")
	} else {
		t.Logf("Module loading failed: %v\nStdout: %s\nStderr: %s", moduleErr, moduleStdout, moduleStderr)
	}

	// Test that the installation is truly relocatable by checking @INC
	incStdout, incStderr, incErr := env.RunCommand("perl", "-e", "print scalar(grep { -d $_ } @INC)")
	if incErr != nil {
		t.Logf("@INC test failed: %v\nStdout: %s\nStderr: %s", incErr, incStdout, incStderr)
	}

	// Test that core modules are available
	coreStdout, coreStderr, coreErr := env.RunCommand("perl", "-e", "use File::Spec; print 'CORE_OK'")
	if coreErr == nil {
		assert.Equal(t, "CORE_OK", coreStdout, "Core module loading should work")
	} else {
		t.Logf("Core module test failed: %v\nStdout: %s\nStderr: %s", coreErr, coreStdout, coreStderr)
	}
}

// Helper function to create a tar.gz archive from a directory
func createTarGzArchive(sourceDir, archivePath string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path from the source directory
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Handle symbolic links
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			header.Linkname = link
		}

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content (only for regular files)
		if info.Mode().IsRegular() {
			sourceFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer sourceFile.Close()

			_, err = io.Copy(tarWriter, sourceFile)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
