package loader

import (
	"brc20query/lib/brc20_swap/model"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var SwapDB *sql.DB

const (
	pg_host     = "10.16.11.95"
	pg_port     = 5432
	pg_user     = "postgres"
	pg_password = "postgres"
	pg_dbname   = "postgres"
)

func Init(psqlInfo string) {
	psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", pg_host, pg_port, pg_user, pg_password, pg_dbname)
	var err error
	SwapDB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Connect PG Failed: ", err)
	}

	SwapDB.SetMaxOpenConns(2000)
	SwapDB.SetMaxIdleConns(1000)
}

// brc20_ticker_info
func SaveDataToDBTickerInfoMap(height int,
	inscriptionsTickerInfoMap map[string]*model.BRC20TokenInfo,
) {
	stmtTickerInfo, err := SwapDB.Prepare(`
INSERT INTO brc20_ticker_info(block_height, tick, max_supply, decimals, limit_per_mint, remaining_supply, pkscript_deployer)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}
	for _, info := range inscriptionsTickerInfoMap {
		// save ticker info
		res, err := stmtTickerInfo.Exec(height, info.Ticker,
			info.Deploy.Max.String(),
			info.Deploy.Decimal,
			info.Deploy.Limit.String(),
			info.Deploy.Max.Sub(info.Deploy.TotalMinted).String(),
			info.Deploy.PkScript,
		)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}
}

func SaveDataToDBTickerBalanceMap(height int,
	tokenUsersBalanceData map[string]map[string]*model.BRC20TokenBalance,
) {
	stmtUserBalance, err := SwapDB.Prepare(`
INSERT INTO brc20_user_balance(block_height, tick, pkscript, available_balance, transferable_balance)
VALUES ($1, $2, $3, $4, $5)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}

	for ticker, holdersMap := range tokenUsersBalanceData {
		// holders
		for _, balanceData := range holdersMap {
			// save balance db
			res, err := stmtUserBalance.Exec(height, ticker,
				balanceData.PkScript,
				balanceData.AvailableBalance.String(),
				balanceData.TransferableBalance.String(),
			)
			if err != nil {
				log.Fatal("PG Statements Exec Wrong: ", err)
			}
			id, err := res.RowsAffected()
			if err != nil {
				log.Fatal("PG Affecte Wrong: ", err)
			}
			fmt.Println(id)
		}
	}
}

func SaveDataToDBTickerHistoryMap(height int,
	inscriptionsTickerInfoMap map[string]*model.BRC20TokenInfo,
) {
	stmtBRC20History, err := SwapDB.Prepare(`
INSERT INTO brc20_history(block_height, tick,
    history_type,
    valid,
    txid,
    idx,
    vout,
    output_value,
    output_offset,
    pkscript_from,
    pkscript_to,
    fee,
    txidx,
    block_time,
    inscription_number,
    inscription_id,
    inscription_content,
	 amount,
	 available_balance,
	 transferable_balance) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}

	for _, info := range inscriptionsTickerInfoMap {
		nValid := 0
		for _, h := range info.History {
			if h.Valid {
				nValid++
			}
		}

		// history
		for _, h := range info.History {
			if !h.Valid {
				continue
			}

			{
				res, err := stmtBRC20History.Exec(height, info.Ticker,
					h.Type, h.Valid,
					h.TxId, h.Idx, h.Vout, h.Satoshi, h.Offset,
					h.PkScriptFrom, h.PkScriptTo,
					h.Fee,
					h.TxIdx, h.BlockTime,
					h.Inscription.InscriptionNumber, h.Inscription.InscriptionId,
					"", // content
					h.Amount, h.AvailableBalance, h.TransferableBalance,
				)
				if err != nil {
					log.Fatal("PG Statements Exec Wrong: ", err)
				}
				id, err := res.RowsAffected()
				if err != nil {
					log.Fatal("PG Affecte Wrong: ", err)
				}
				fmt.Println(id)
			}
		}
	}
}

func SaveDataToDBTransferStateMap(height int,
	inscriptionsTransferRemoveMap map[string]struct{},
) {
	stmtTransferState, err := SwapDB.Prepare(`
INSERT INTO brc20_transfer_state(block_height, create_key, moved)
VALUES ($1, $2, $3)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}

	for createKey := range inscriptionsTransferRemoveMap {
		res, err := stmtTransferState.Exec(height, createKey, true)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}
}

func SaveDataToDBValidTransferMap(height int,
	inscriptionsValidTransferMap map[string]*model.InscriptionBRC20TickInfo,
) {
	stmtValidTransfer, err := SwapDB.Prepare(`
INSERT INTO brc20_valid_transfer(block_height, tick, pkscript, amount,
    inscription_number, inscription_id, txid, vout, output_value, output_offset)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}

	for _, transferInfo := range inscriptionsValidTransferMap {
		res, err := stmtValidTransfer.Exec(height, transferInfo.Tick,
			transferInfo.PkScript,
			transferInfo.Amount.String(),
			transferInfo.InscriptionNumber, transferInfo.Meta.GetInscriptionId(),
			transferInfo.TxId, transferInfo.Vout, transferInfo.Satoshi, transferInfo.Offset,
		)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}

}

func SaveDataToDBModuleInfoMap(height int,
	modulesInfoMap map[string]*model.BRC20ModuleSwapInfo) {

	stmtSwapInfo, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_info(block_height, module_id,
	 name,
    pkscript_deployer,
    pkscript_sequencer,
    pkscript_gas_to,
    pkscript_lp_fee,
	 gas_tick,
    fee_rate_swap
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}

	for moduleId, info := range modulesInfoMap {
		// save swap info db
		res, err := stmtSwapInfo.Exec(height, moduleId,
			info.Name,
			info.DeployerPkScript,
			info.SequencerPkScript,
			info.GasToPkScript,
			info.LpFeePkScript,
			info.GasTick,
			info.FeeRateSwap,
		)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}
}

func SaveDataToDBModuleHistoryMap(height int,
	modulesInfoMap map[string]*model.BRC20ModuleSwapInfo) {

	stmtSwapHistory, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_history(block_height, module_id,
    history_type,
    valid,
    txid,
    idx,
    vout,
    output_value,
    output_offset,
    pkscript_from,
    pkscript_to,
    fee,
    txidx,
    block_time,
    inscription_number,
    inscription_id,
    inscription_content
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}

	for moduleId, info := range modulesInfoMap {

		nValid := 0
		// history
		for _, h := range info.History {
			if h.Valid {
				nValid++
			}
			if !h.Valid {
				continue
			}

			{
				res, err := stmtSwapHistory.Exec(height, moduleId,
					h.Type, h.Valid,
					h.TxId, h.Idx, h.Vout, h.Satoshi, h.Offset,
					h.PkScriptFrom, h.PkScriptTo,
					h.Fee,
					h.TxIdx, h.BlockTime,
					h.Inscription.InscriptionNumber, h.Inscription.InscriptionId,
					"", // content
				)
				if err != nil {
					log.Fatal("PG Statements Exec Wrong: ", err)
				}
				id, err := res.RowsAffected()
				if err != nil {
					log.Fatal("PG Affecte Wrong: ", err)
				}
				fmt.Println(id)
			}

		}

	}
}

// approve
func SaveDataToDBSwapApproveStateMap(height int,
	inscriptionsApproveRemoveMap map[string]struct{},
) {
	stmtApproveState, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_approve_state(block_height, create_key, moved)
VALUES ($1, $2, $3)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}
	for createKey := range inscriptionsApproveRemoveMap {
		res, err := stmtApproveState.Exec(height, createKey, true)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}
}

func SaveDataToDBSwapApproveMap(height int,
	inscriptionsValidApproveMap map[string]*model.InscriptionBRC20SwapInfo,
) {
	stmtValidApprove, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_valid_approve(block_height, module_id, tick, pkscript, amount,
    inscription_number, inscription_id, txid, vout, output_value, output_offset)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}
	for _, approveInfo := range inscriptionsValidApproveMap {
		res, err := stmtValidApprove.Exec(height, approveInfo.Module, approveInfo.Tick,
			approveInfo.Data.PkScript,
			approveInfo.Amount.String(),
			approveInfo.Data.InscriptionNumber, approveInfo.Data.GetInscriptionId(),
			approveInfo.Data.TxId, approveInfo.Data.Vout, approveInfo.Data.Satoshi, approveInfo.Data.Offset,
		)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}
}

// cond approve
func SaveDataToDBSwapCondApproveStateMap(height int,
	inscriptionsCondApproveRemoveMap map[string]struct{},
) {
	stmtCondApproveState, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_cond_approve_state(block_height, create_key, moved)
VALUES ($1, $2, $3)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}
	for createKey := range inscriptionsCondApproveRemoveMap {
		res, err := stmtCondApproveState.Exec(height, createKey, true)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}
}

func SaveDataToDBSwapCondApproveMap(height int,
	inscriptionsValidConditionalApproveMap map[string]*model.InscriptionBRC20SwapConditionalApproveInfo,
) {
	stmtValidCondApprove, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_valid_cond_approve(block_height, module_id, tick, pkscript, amount,
    inscription_number, inscription_id, txid, vout, output_value, output_offset)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}
	for _, condApproveInfo := range inscriptionsValidConditionalApproveMap {
		res, err := stmtValidCondApprove.Exec(height, condApproveInfo.Module, condApproveInfo.Tick,
			condApproveInfo.Data.PkScript,
			condApproveInfo.Amount.String(),
			condApproveInfo.Data.InscriptionNumber, condApproveInfo.Data.GetInscriptionId(),
			condApproveInfo.Data.TxId, condApproveInfo.Data.Vout, condApproveInfo.Data.Satoshi, condApproveInfo.Data.Offset,
		)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}
}

// withdraw
func SaveDataToDBSwapWithdrawStateMap(height int,
	inscriptionsWithdrawRemoveMap map[string]struct{},
) {
	stmtWithdrawState, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_withdraw_state(block_height, create_key, moved)
VALUES ($1, $2, $3)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}

	for createKey := range inscriptionsWithdrawRemoveMap {
		res, err := stmtWithdrawState.Exec(height, createKey, true)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}
}

func SaveDataToDBSwapWithdrawMap(height int,
	inscriptionsValidWithdrawMap map[string]*model.InscriptionBRC20SwapInfo,
) {
	stmtValidWithdraw, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_valid_withdraw(block_height, module_id, tick, pkscript, amount,
    inscription_number, inscription_id, txid, vout, output_value, output_offset)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}

	for _, withdrawInfo := range inscriptionsValidWithdrawMap {
		res, err := stmtValidWithdraw.Exec(height, withdrawInfo.Module, withdrawInfo.Tick,
			withdrawInfo.Data.PkScript,
			withdrawInfo.Amount.String(),
			withdrawInfo.Data.InscriptionNumber, withdrawInfo.Data.GetInscriptionId(),
			withdrawInfo.Data.TxId, withdrawInfo.Data.Vout, withdrawInfo.Data.Satoshi, withdrawInfo.Data.Offset,
		)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}

}

// commit
func SaveDataToDBSwapCommitStateMap(height int,
	inscriptionsCommitRemoveMap map[string]struct{},
) {
	stmtCommitState, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_commit_state(block_height, create_key, moved)
VALUES ($1, $2, $3)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}
	for createKey := range inscriptionsCommitRemoveMap {
		res, err := stmtCommitState.Exec(height, createKey, true)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}
}

func SaveDataToDBSwapCommitMap(height int,
	inscriptionsValidCommitMap map[string]*model.InscriptionBRC20Data,
) {
	stmtValidCommit, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_commit_info(block_height, module_id, pkscript,
    inscription_number, inscription_id, txid, vout, output_value, output_offset, inscription_content)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}
	for _, commitInfo := range inscriptionsValidCommitMap {
		res, err := stmtValidCommit.Exec(height, "commitInfo.Module", commitInfo.PkScript,
			commitInfo.InscriptionNumber, commitInfo.GetInscriptionId(),
			commitInfo.TxId, commitInfo.Vout, commitInfo.Satoshi, commitInfo.Offset, commitInfo.ContentBody,
		)
		if err != nil {
			log.Fatal("PG Statements Exec Wrong: ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			log.Fatal("PG Affecte Wrong: ", err)
		}
		fmt.Println(id)
	}
}

func SaveDataToDBModuleCommitChainMap(height int,
	modulesInfoMap map[string]*model.BRC20ModuleSwapInfo) {
	stmtSwapCommitChain, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_commit_chain(block_height, module_id, commit_id, valid, connected)
VALUES ($1, $2, $3, $4, $5)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}

	for moduleId, info := range modulesInfoMap {
		// commit state
		commitState := make(map[string]*[2]bool)
		for commitId := range info.CommitInvalidMap {
			if state, ok := commitState[commitId]; !ok {
				commitState[commitId] = &[2]bool{false, false}
			} else {
				state[0] = false
			}
		}

		for commitId := range info.CommitIdMap {
			if state, ok := commitState[commitId]; !ok {
				commitState[commitId] = &[2]bool{true, false}
			} else {
				state[0] = true
			}
		}
		for commitId := range info.CommitIdChainMap {
			if state, ok := commitState[commitId]; !ok {
				commitState[commitId] = &[2]bool{true, true}
			} else {
				state[1] = true
			}
		}

		// save commit state db
		for commitId, state := range commitState {
			res, err := stmtSwapCommitChain.Exec(height, moduleId,
				commitId,
				state[0], // valid
				state[1], // connected
			)
			if err != nil {
				log.Fatal("PG Statements Exec Wrong: ", err)
			}
			id, err := res.RowsAffected()
			if err != nil {
				log.Fatal("PG Affecte Wrong: ", err)
			}
			fmt.Println(id)
		}
	}
}

func SaveDataToDBModuleUserBalanceMap(height int,
	modulesInfoMap map[string]*model.BRC20ModuleSwapInfo) {

	stmtUserBalance, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_user_balance(block_height, module_id, tick,
    pkscript, swap_balance, available_balance, approveable_balance, cond_approveable_balance, withdrawable_balance)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}

	for moduleId, info := range modulesInfoMap {
		for ticker, holdersMap := range info.TokenUsersBalanceDataMap {
			// holders
			for _, balanceData := range holdersMap {
				// save balance db
				res, err := stmtUserBalance.Exec(height, moduleId, ticker,
					balanceData.PkScript,
					balanceData.SwapAccountBalance.String(),
					balanceData.AvailableBalance.String(),
					balanceData.ApproveableBalance.String(),
					balanceData.CondApproveableBalance.String(),
					balanceData.WithdrawableBalance.String(),
				)
				if err != nil {
					log.Fatal("PG Statements Exec Wrong: ", err)
				}
				id, err := res.RowsAffected()
				if err != nil {
					log.Fatal("PG Affecte Wrong: ", err)
				}
				fmt.Println(id)
			}
		}
	}
}

func SaveDataToDBModulePoolLpBalanceMap(height int,
	modulesInfoMap map[string]*model.BRC20ModuleSwapInfo) {

	stmtPoolBalance, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_pool_balance(block_height, module_id, tick0, tick0_balance, tick1, tick1_balance, lp_balance)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}
	for moduleId, info := range modulesInfoMap {
		for _, swap := range info.SwapPoolTotalBalanceDataMap {
			// save swap balance db
			res, err := stmtPoolBalance.Exec(height, moduleId,
				swap.Tick[0],
				swap.TickBalance[0],
				swap.Tick[1],
				swap.TickBalance[1],
				swap.LpBalance.String(),
			)
			if err != nil {
				log.Fatal("PG Statements Exec Wrong: ", err)
			}
			id, err := res.RowsAffected()
			if err != nil {
				log.Fatal("PG Affecte Wrong: ", err)
			}
			fmt.Println(id)
		}
	}
}

func SaveDataToDBModuleUserLpBalanceMap(height int,
	modulesInfoMap map[string]*model.BRC20ModuleSwapInfo) {

	stmtLpBalance, err := SwapDB.Prepare(`
INSERT INTO brc20_swap_user_lp_balance(block_height, module_id, pool, pkscript, lp_balance)
VALUES ($1, $2, $3, $4, $5)
`)
	if err != nil {
		log.Fatal("PG Statements Wrong: ", err)
	}
	for moduleId, info := range modulesInfoMap {
		for ticker, holdersMap := range info.LPTokenUsersBalanceMap {
			// holders
			for holder, balanceData := range holdersMap {
				// save balance db
				res, err := stmtLpBalance.Exec(height, moduleId, ticker,
					holder,
					balanceData.String(),
				)
				if err != nil {
					log.Fatal("PG Statements Exec Wrong: ", err)
				}
				id, err := res.RowsAffected()
				if err != nil {
					log.Fatal("PG Affecte Wrong: ", err)
				}
				fmt.Println(id)
			}
		}
	}

}

func SaveDataToDBModuleTickInfoMap(moduleId string, condStateBalanceDataMap map[string]*model.BRC20ModuleConditionalApproveStateBalance,
	inscriptionsTickerInfoMap, userTokensBalanceData map[string]map[string]*model.BRC20ModuleTokenBalance) {

	// condStateBalanceDataMap
	for ticker, stateBalance := range condStateBalanceDataMap {
		fmt.Printf("  module deposit/withdraw state: %s deposit: %s, match: %s, new: %s, cancel: %s, wait: %s\n",
			ticker,
			stateBalance.BalanceDeposite.String(),
			stateBalance.BalanceApprove.String(),
			stateBalance.BalanceNewApprove.String(),
			stateBalance.BalanceCancelApprove.String(),

			stateBalance.BalanceNewApprove.Sub(
				stateBalance.BalanceApprove).Sub(
				stateBalance.BalanceCancelApprove).String(),
		)
	}

	fmt.Printf("\n")
}
