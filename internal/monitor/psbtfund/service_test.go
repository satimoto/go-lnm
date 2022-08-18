package psbtfund_test

import (
	"os"
	"time"

	"testing"

	"github.com/satimoto/go-datastore/pkg/util"
)

func routine(t *testing.T, expiry time.Time) {
	until := time.Until(expiry)

	t.Errorf("PSBT Fund timeout (%v) set for %v: %v", time.Now(), expiry, until)
}

func TestService(t *testing.T) {
	os.Setenv("PSBT_BATCH_TIMEOUT", "30")
	defer os.Unsetenv("PSBT_BATCH_TIMEOUT")

	t.Run("No session invoice", func(t *testing.T) {
		psbtBatchTimeout := util.GetEnvInt32("PSBT_BATCH_TIMEOUT", 30)
		expiry := time.Now().Add(time.Duration(psbtBatchTimeout) * time.Second)
		
		go routine(t, expiry)

		time.Sleep(1 * time.Second)
	})
}
