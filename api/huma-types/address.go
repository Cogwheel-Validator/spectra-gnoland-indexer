package humatypes

import (
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
)

type AddressGetInput struct {
	Address       string    `path:"address" doc:"Address" required:"true"`
	FromTimestamp time.Time `query:"from_timestamp" doc:"From timestamp" format:"2025-01-01T00:00:00+00:00 or 2025-01-01T00:00:00Z" required:"true"`
	ToTimestamp   time.Time `query:"to_timestamp" doc:"To timestamp" format:"2025-01-01T00:00:00+00:00 or 2025-01-01T00:00:00Z" required:"true"`
}

type AddressGetOutput struct {
	Body []database.AddressTx
}
