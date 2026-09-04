package settlement

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fluxa/fluxa/internal/queue"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// TransferSubmitter is the engine surface the worker needs. Both the
// Stellar Engine and the XDCEngine implement it; the worker picks one at
// startup based on CHAIN_BACKEND.
type TransferSubmitter interface {
	SubmitTransfer(ctx context.Context, txID string) error
}

type Worker struct {
	engine TransferSubmitter
}

func NewWorker(engine TransferSubmitter) *Worker {
	return &Worker{engine: engine}
}

func (w *Worker) HandleProcessTransfer(ctx context.Context, task *asynq.Task) error {
	var payload queue.ProcessTransferPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	log.Info().Str("tx_id", payload.TransactionID).Msg("processing transfer")

	if err := w.engine.SubmitTransfer(ctx, payload.TransactionID); err != nil {
		log.Error().Err(err).Str("tx_id", payload.TransactionID).Msg("transfer submission failed")
		return err
	}

	log.Info().Str("tx_id", payload.TransactionID).Msg("transfer confirmed")
	return nil
}
