package preview

import (
	"os"

	tlog "github.com/aquasecurity/trivy/pkg/log"
)

func init() {
	ll := tlog.New(tlog.NewHandler(os.Stderr, &tlog.Options{
		Level: tlog.LevelDebug,
	}))
	var _ = ll
	//tlog.SetDefault(ll)
}
