package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("enable-feature-flag command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("enable-feature-flag", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("enable-feature-flag - Allow use of a feature"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf enable-feature-flag FEATURE_NAME"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("disable-feature-flag, feature-flags"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})