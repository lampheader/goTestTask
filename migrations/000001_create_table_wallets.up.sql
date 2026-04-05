CREATE TABLE IF NOT EXISTS wallets (
    uuid uuid DEFAULT gen_random_uuid() not null PRIMARY KEY,
    balance double precision default 0 NOT NULL,
    version integer DEFAULT 1 NOT NULL 
    );

ALTER TABLE wallets 
    ADD CONSTRAINT balance_non_negative_check
        CHECK (balance >= 0);