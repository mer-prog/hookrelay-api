CREATE INDEX idx_delivery_logs_endpoint_created ON delivery_logs (endpoint_id, created_at DESC);
CREATE INDEX idx_delivery_logs_status_retry ON delivery_logs (status, next_retry_at);
CREATE INDEX idx_delivery_logs_event ON delivery_logs (event_id);
CREATE INDEX idx_endpoints_user_active ON endpoints (user_id, is_active);
CREATE INDEX idx_api_keys_hash ON api_keys (key_hash);
