drop table if exists ledger;
create table ledger (
    ledger_id varchar(50) not null primary key,
    name varchar(50) not null,
    is_subledger bool,
    parent_ledger_id varchar(50)

);
create index idx_ledger_parent_ledger_id on ledger(parent_ledger_id);

drop table if exists account;
create table account (
    account_id varchar(50) not null primary key,
    ledger_id varchar(50) not null,
    class varchar(25) not null,
    code varchar(25) not null,
    name varchar(50) not null,
    metadata jsonb,
    parent_account_id varchar(50),
    currency varchar(10)
);
create index idx_account_ledger_id on account(ledger_id);
create index idx_account_code on account(code);
create index idx_account_parent_account_id on account(parent_account_id);
create unique index idx_unique_ledger_code on account(ledger_id, code);
create index idx_metadata_gin on account using gin (metadata);

drop table if exists account_balance;
create table account_balance (
    account_id varchar(50) not null,
    balance_as_of_transaction_id varchar(50) not null,
    balance integer not null,
    balance_date timestamptz,
    PRIMARY KEY (account_id, balance_as_of_transaction_id)
);
create index idx_account_balance_balance_date on account_balance(balance_date);

drop table if exists ledger_transaction;
create table ledger_transaction (
    ledger_transaction_id varchar(50) not null primary key,
    ledger_id varchar(50) not null,
    metadata jsonb,
    transaction_date timestamptz
);
create index idx_ledger_transaction_tx_date on ledger_transaction(transaction_date);
create index idx_ledger_transaction_ledget_id on ledger_transaction(ledger_id);

drop table if exists ledger_transaction_entry;
create table ledger_transaction_entry (
    ledger_transaction_id varchar(50) not null,
    account_id varchar(50) not null,
    transaction_entry_type varchar(10),
    amount integer not null,
    metadata jsonb,
    currency varchar(10),
    PRIMARY KEY (ledger_transaction_id, account_id)
);

