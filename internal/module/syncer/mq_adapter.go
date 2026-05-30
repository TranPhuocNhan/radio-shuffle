package syncer

import "github.com/tranphuocnhan/radio-shuffle/internal/platform/mq"

// toSyncRequestedMessage maps a domain command to an MQ payload.
func toSyncRequestedMessage(cmd SyncCommand) mq.SyncRequestedMessage {
	return mq.SyncRequestedMessage{
		RequestID:   cmd.RequestID,
		RequestedBy: cmd.RequestedBy,
		Scope:       cmd.Scope,
		RequestedAt: cmd.RequestedAt,
	}
}

// FromSyncRequestedMessage maps an MQ payload to a domain command.
func FromSyncRequestedMessage(msg mq.SyncRequestedMessage) SyncCommand {
	return SyncCommand{
		RequestID:   msg.RequestID,
		RequestedBy: msg.RequestedBy,
		Scope:       msg.Scope,
		RequestedAt: msg.RequestedAt,
	}
}
