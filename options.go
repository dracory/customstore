package customstore

// RecordOption represents a functional option that mutates a RecordInterface
// instance during construction or afterwards.
type RecordOption func(RecordInterface) error

// WithID sets the record ID.
func WithID(id string) RecordOption {
	return func(r RecordInterface) error {
		r.SetID(id)
		return nil
	}
}

// WithMemo sets the record memo.
func WithMemo(memo string) RecordOption {
	return func(r RecordInterface) error {
		r.SetMemo(memo)
		return nil
	}
}

// WithMetas sets the record metas (overwrites existing metas).
func WithMetas(metas map[string]string) RecordOption {
	return func(r RecordInterface) error {
		return r.SetMetas(metas)
	}
}

// WithPayload sets the record payload (raw JSON string).
func WithPayload(payload string) RecordOption {
	return func(r RecordInterface) error {
		r.SetPayload(payload)
		return nil
	}
}

// WithPayloadMap sets the record payload from a map (will be marshaled to JSON).
func WithPayloadMap(payloadMap map[string]any) RecordOption {
	return func(r RecordInterface) error {
		return r.SetPayloadMap(payloadMap)
	}
}
