CREATE TABLE inbox (
    message_id VARCHAR(255) NOT NULL,
    consumer_name VARCHAR(100) NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (message_id, consumer_name)
);

CREATE TABLE outbox (
    event_id VARCHAR(36) PRIMARY KEY,
    aggregate_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);