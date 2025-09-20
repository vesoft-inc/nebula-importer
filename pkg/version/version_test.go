package version

import (
	"fmt"
	"runtime"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	var v *Version

	BeforeEach(func() {
		buildVersionPatches := gomonkey.ApplyGlobalVar(&buildVersion, "buildVersion")
		defer buildVersionPatches.Reset()
		buildCommitPatches := gomonkey.ApplyGlobalVar(&buildCommit, "buildCommit")
		defer buildCommitPatches.Reset()
		buildCommitDatePatches := gomonkey.ApplyGlobalVar(&buildCommitDate, "buildCommitDate")
		defer buildCommitDatePatches.Reset()
		buildDatePatches := gomonkey.ApplyGlobalVar(&buildDate, "buildDate")
		defer buildDatePatches.Reset()

		v = GetVersion()
	})
	Describe("GetVersion", func() {
		It("should be", func() {
			Expect(v).Should(Equal(&Version{
				Version:    "buildVersion",
				Commit:     "buildCommit",
				CommitDate: "buildCommitDate",
				BuildDate:  "buildDate",
				GoVersion:  runtime.Version(),
				Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			}))
		})
	})

	Describe(".String", func() {
		It("should be", func() {
			Expect(v.String()).Should(Equal(
				fmt.Sprintf(`Version:    buildVersion
Commit:     buildCommit
CommitDate: buildCommitDate
BuildDate:  buildDate
GoVersion:  %s
Platform:   %s
`, runtime.Version(), fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)),
			))
		})
	})
})
