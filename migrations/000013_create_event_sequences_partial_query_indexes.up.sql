CREATE index idx_balances_transfer_accountid_source on event_sequences ((data->0->>'value'))
 where method='Transfer' and section='balances';

CREATE index idx_balances_transfer_accountid_dest on event_sequences ((data->1->>'value'))
 where method='Transfer' and section='balances';

CREATE index idx_balances_deposit_accountid on event_sequences ((data->0->>'value'))
 where method='Deposit' and section='balances';

CREATE index idx_staking_bonded_accountid on event_sequences ((data->0->>'value'))
 where method='Bonded' and section='staking';

CREATE index idx_staking_unbonded_accountid on event_sequences ((data->0->>'value'))
 where method='Unbonded' and section='staking';

CREATE index idx_staking_withdrawn_accountid on event_sequences ((data->0->>'value'))
 where method='Withdrawn' and section='staking';