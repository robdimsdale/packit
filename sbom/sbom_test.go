package sbom_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSBOM(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Generate", func() {
		context("when given a directory", func() {
			it("generates an SBOM for that directory", func() {
				bom, err := sbom.Generate("testdata/")
				Expect(err).NotTo(HaveOccurred())

				formatter, err := bom.InFormats(sbom.SyftFormat)
				Expect(err).NotTo(HaveOccurred())

				syft := bytes.NewBuffer(nil)
				_, err = io.Copy(syft, formatter.Formats()[0].Content)
				Expect(err).NotTo(HaveOccurred())

				var syftOutput syftOutput
				err = json.Unmarshal(syft.Bytes(), &syftOutput)
				Expect(err).NotTo(HaveOccurred(), syft.String())
				Expect(syftOutput.Source.Type).To(Equal("directory"), syft.String())
			})
		})

		context("when given a file", func() {
			it("generates an SBOM for that file", func() {
				bom, err := sbom.Generate("testdata/package-lock.json")
				Expect(err).NotTo(HaveOccurred())

				formatter, err := bom.InFormats(sbom.SyftFormat)
				Expect(err).NotTo(HaveOccurred())

				syft := bytes.NewBuffer(nil)
				_, err = io.Copy(syft, formatter.Formats()[0].Content)
				Expect(err).NotTo(HaveOccurred())

				var syftOutput syftOutput
				err = json.Unmarshal(syft.Bytes(), &syftOutput)
				Expect(err).NotTo(HaveOccurred(), syft.String())
				Expect(syftOutput.Source.Type).To(Equal("file"), syft.String())
			})
		})

		context("failure cases", func() {
			context("when given a nonexistent path", func() {
				it("returns an error", func() {
					_, err := sbom.Generate("no/such/path")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})

	context("GenerateFromDependency", func() {
		it("generates a SBOM from a dependency for latest schema versions", func() {
			bom, err := sbom.GenerateFromDependency(postal.Dependency{
				CPE:          "cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*",
				ID:           "go",
				Licenses:     []string{"BSD-3-Clause"},
				Name:         "Go",
				PURL:         "pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz",
				SHA256:       "ca9ef23a5db944b116102b87c1ae9344b27e011dae7157d2f1e501abd39e9829",
				Source:       "https://dl.google.com/go/go1.16.9.src.tar.gz",
				SourceSHA256: "0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d",
				Stacks:       []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny"},
				URI:          "https://deps.paketo.io/go/go_go1.16.9_linux_x64_bionic_ca9ef23a.tgz",
				Version:      "1.16.9",
			}, "some-path")
			Expect(err).NotTo(HaveOccurred())

			formatter, err := bom.InFormats(sbom.SyftFormat, sbom.CycloneDXFormat, sbom.SPDXFormat)
			Expect(err).NotTo(HaveOccurred())

			formats := formatter.Formats()

			syft := bytes.NewBuffer(nil)
			for _, format := range formats {
				if format.Extension == "syft.json" {
					_, err = io.Copy(syft, format.Content)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			var syftLatestOutput syftOutput

			err = json.Unmarshal(syft.Bytes(), &syftLatestOutput)
			Expect(err).NotTo(HaveOccurred(), syft.String())

			Expect(syftLatestOutput.Schema.Version).To(MatchRegexp(`3\.\d+\.\d+`), syft.String())

			goArtifact := syftLatestOutput.Artifacts[0]
			Expect(goArtifact.Name).To(Equal("Go"), syft.String())
			Expect(goArtifact.Version).To(Equal("1.16.9"), syft.String())
			Expect(goArtifact.Licenses).To(Equal([]string{"BSD-3-Clause"}), syft.String())
			Expect(goArtifact.CPEs).To(Equal([]string{"cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*"}), syft.String())
			Expect(goArtifact.PURL).To(Equal("pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz"), syft.String())
			Expect(syftLatestOutput.Source.Type).To(Equal("directory"), syft.String())
			Expect(syftLatestOutput.Source.Target).To(Equal("some-path"), syft.String())

			cdx := bytes.NewBuffer(nil)
			for _, format := range formats {
				if format.Extension == "cdx.json" {
					_, err = io.Copy(cdx, format.Content)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			var cdxLatestOutput cdxOutput

			err = json.Unmarshal(cdx.Bytes(), &cdxLatestOutput)
			Expect(err).NotTo(HaveOccurred(), cdx.String())

			Expect(cdxLatestOutput.BOMFormat).To(Equal("CycloneDX"))
			Expect(cdxLatestOutput.SpecVersion).To(Equal("1.4"))

			goComponent := cdxLatestOutput.Components[0]
			Expect(goComponent.Name).To(Equal("Go"), cdx.String())
			Expect(goComponent.Version).To(Equal("1.16.9"), cdx.String())
			Expect(goComponent.Licenses).To(HaveLen(1), cdx.String())
			Expect(goComponent.Licenses[0].License.ID).To(Equal("BSD-3-Clause"), cdx.String())
			Expect(goComponent.PURL).To(Equal("pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz"), cdx.String())

			Expect(cdxLatestOutput.Metadata.Component.Type).To(Equal("file"), cdx.String())
			Expect(cdxLatestOutput.Metadata.Component.Name).To(Equal("some-path"), cdx.String())

			spdx := bytes.NewBuffer(nil)
			for _, format := range formats {
				if format.Extension == "spdx.json" {
					_, err = io.Copy(spdx, format.Content)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			var spdxLatestOutput spdxOutput

			err = json.Unmarshal(spdx.Bytes(), &spdxLatestOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(err).NotTo(HaveOccurred(), spdx.String())

			Expect(spdxLatestOutput.SPDXVersion).To(Equal("SPDX-2.2"), spdx.String())

			goPackage := spdxLatestOutput.Packages[0]
			Expect(goPackage.Name).To(Equal("Go"), spdx.String())
			Expect(goPackage.Version).To(Equal("1.16.9"), spdx.String())
			Expect(goPackage.LicenseConcluded).To(Equal("BSD-3-Clause"), spdx.String())
			Expect(goPackage.LicenseDeclared).To(Equal("BSD-3-Clause"), spdx.String())
			Expect(goPackage.ExternalRefs).To(ContainElement(externalRef{
				Category: "SECURITY",
				Locator:  "cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*",
				Type:     "cpe23Type",
			}), spdx.String())
			Expect(goPackage.ExternalRefs).To(ContainElement(externalRef{
				Category: "PACKAGE_MANAGER",
				Locator:  "pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz",
				Type:     "purl",
			}), spdx.String())
		})

		it("generates a SBOM from a dependency as syft2 JSON", func() {
			bom, err := sbom.GenerateFromDependency(postal.Dependency{
				CPE:          "cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*",
				ID:           "go",
				Licenses:     []string{"BSD-3-Clause"},
				Name:         "Go",
				PURL:         "pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz",
				SHA256:       "ca9ef23a5db944b116102b87c1ae9344b27e011dae7157d2f1e501abd39e9829",
				Source:       "https://dl.google.com/go/go1.16.9.src.tar.gz",
				SourceSHA256: "0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d",
				Stacks:       []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny"},
				URI:          "https://deps.paketo.io/go/go_go1.16.9_linux_x64_bionic_ca9ef23a.tgz",
				Version:      "1.16.9",
			}, "some-path")
			Expect(err).NotTo(HaveOccurred())

			formatter, err := bom.InFormats(fmt.Sprintf("%s;version=2.0.2", sbom.SyftFormat))
			Expect(err).NotTo(HaveOccurred())

			formats := formatter.Formats()

			syft := bytes.NewBuffer(nil)
			for _, format := range formats {
				if format.Extension == "syft.json" {
					_, err = io.Copy(syft, format.Content)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			var syft2Output syftOutput

			err = json.Unmarshal(syft.Bytes(), &syft2Output)
			Expect(err).NotTo(HaveOccurred(), syft.String())

			Expect(syft2Output.Schema.Version).To(Equal("2.0.2"), syft.String())

			goArtifact := syft2Output.Artifacts[0]
			Expect(goArtifact.Name).To(Equal("Go"), syft.String())
			Expect(goArtifact.Version).To(Equal("1.16.9"), syft.String())
			Expect(goArtifact.Licenses).To(Equal([]string{"BSD-3-Clause"}), syft.String())
			Expect(goArtifact.CPEs).To(Equal([]string{"cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*"}), syft.String())
			Expect(goArtifact.PURL).To(Equal("pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz"), syft.String())
			Expect(syft2Output.Source.Type).To(Equal("directory"), syft.String())
			Expect(syft2Output.Source.Target).To(Equal("some-path"), syft.String())
		})

		it("generates a SBOM from a dependency in CycloneDX 1.3 JSON", func() {
			bom, err := sbom.GenerateFromDependency(postal.Dependency{
				CPE:          "cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*",
				ID:           "go",
				Licenses:     []string{"BSD-3-Clause"},
				Name:         "Go",
				PURL:         "pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz",
				SHA256:       "ca9ef23a5db944b116102b87c1ae9344b27e011dae7157d2f1e501abd39e9829",
				Source:       "https://dl.google.com/go/go1.16.9.src.tar.gz",
				SourceSHA256: "0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d",
				Stacks:       []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny"},
				URI:          "https://deps.paketo.io/go/go_go1.16.9_linux_x64_bionic_ca9ef23a.tgz",
				Version:      "1.16.9",
			}, "some-path")
			Expect(err).NotTo(HaveOccurred())

			formatter, err := bom.InFormats(fmt.Sprintf("%s;version=1.3", sbom.CycloneDXFormat))
			Expect(err).NotTo(HaveOccurred())

			formats := formatter.Formats()

			cdx := bytes.NewBuffer(nil)
			for _, format := range formats {
				if format.Extension == "cdx.json" {
					_, err = io.Copy(cdx, format.Content)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			var cdx13Output cdxOutput

			err = json.Unmarshal(cdx.Bytes(), &cdx13Output)
			Expect(err).NotTo(HaveOccurred(), cdx.String())

			Expect(cdx13Output.BOMFormat).To(Equal("CycloneDX"))
			Expect(cdx13Output.SpecVersion).To(Equal("1.3"))

			goComponent := cdx13Output.Components[0]
			Expect(goComponent.Name).To(Equal("Go"), cdx.String())
			Expect(goComponent.Version).To(Equal("1.16.9"), cdx.String())
			Expect(goComponent.Licenses).To(HaveLen(1), cdx.String())
			Expect(goComponent.Licenses[0].License.ID).To(Equal("BSD-3-Clause"), cdx.String())
			Expect(goComponent.PURL).To(Equal("pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz"), cdx.String())

			Expect(cdx13Output.Metadata.Component.Type).To(Equal("file"), cdx.String())
			Expect(cdx13Output.Metadata.Component.Name).To(Equal("some-path"), cdx.String())
		})

		context("failure cases", func() {
			context("when the CPE is invalid", func() {
				it("returns an error", func() {
					_, err := sbom.GenerateFromDependency(postal.Dependency{
						CPE: "not a valid CPE",
					}, "some-path")
					Expect(err).To(MatchError(ContainSubstring("failed to parse CPE")))
				})
			})
		})
	})

	context("InFormats", func() {
		context("failure cases", func() {
			context("when a format is not supported", func() {
				it("returns an error", func() {
					_, err := sbom.SBOM{}.InFormats("unknown-format")
					Expect(err).To(MatchError(`unsupported SBOM format: 'unknown-format'`))
				})
			})
			context("when a requested version is not supported", func() {
				it("returns an error", func() {
					_, err := sbom.SBOM{}.InFormats(fmt.Sprintf("%s;version=0.0.0", sbom.SyftFormat))
					Expect(err).To(MatchError(fmt.Sprintf(`version '0.0.0' is not supported for SBOM format '%s'`, sbom.SyftFormat)))
				})
			})
		})
	})
}
