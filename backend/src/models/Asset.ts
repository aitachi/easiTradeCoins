//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

import database from '../database';
import Decimal from 'decimal.js';

export interface UserAsset {
  asset_id: number;
  user_id: number;
  currency: string;
  chain: string;
  available: string;
  frozen: string;
  total: string;
  created_at: Date;
  updated_at: Date;
}

export interface Deposit {
  deposit_id: number;
  user_id: number;
  currency: string;
  chain: string;
  amount: string;
  address: string;
  from_address?: string;
  txid: string;
  confirmations: number;
  required_confirmations: number;
  status: number;
  error_reason?: string;
  created_at: Date;
  confirmed_at?: Date;
  updated_at: Date;
}

export interface Withdrawal {
  withdrawal_id: number;
  user_id: number;
  currency: string;
  chain: string;
  amount: string;
  fee: string;
  actual_amount: string;
  address: string;
  address_tag?: string;
  txid?: string;
  status: number;
  audit_user_id?: number;
  audit_time?: Date;
  complete_time?: Date;
  created_at: Date;
  updated_at: Date;
  remark?: string;
  reject_reason?: string;
}

export interface AssetTransaction {
  tx_id: number;
  user_id: number;
  currency: string;
  chain: string;
  amount: string;
  balance_before: string;
  balance_after: string;
  tx_type: number;
  ref_id?: number;
  description?: string;
  created_at: Date;
}

class AssetModel {
  // 获取用户资产
  async getUserAsset(
    userId: number,
    currency: string,
    chain: string
  ): Promise<UserAsset | null> {
    const query = `
      SELECT * FROM user_assets
      WHERE user_id = $1 AND currency = $2 AND chain = $3
    `;
    const result = await database.query(query, [userId, currency, chain]);
    return result.rows[0] || null;
  }

  // 获取用户所有资产
  async getUserAssets(userId: number): Promise<UserAsset[]> {
    const query = `
      SELECT * FROM user_assets
      WHERE user_id = $1
      ORDER BY currency, chain
    `;
    const result = await database.query(query, [userId]);
    return result.rows;
  }

  // 创建或初始化资产
  async ensureAsset(userId: number, currency: string, chain: string): Promise<UserAsset> {
    const existing = await this.getUserAsset(userId, currency, chain);
    if (existing) return existing;

    const query = `
      INSERT INTO user_assets (user_id, currency, chain, available, frozen)
      VALUES ($1, $2, $3, 0, 0)
      RETURNING *
    `;
    const result = await database.query(query, [userId, currency, chain]);
    return result.rows[0];
  }

  // 增加可用余额
  async addAvailable(
    client: any,
    userId: number,
    currency: string,
    chain: string,
    amount: string
  ): Promise<void> {
    await this.ensureAsset(userId, currency, chain);

    const query = `
      UPDATE user_assets
      SET available = available + $1, updated_at = NOW()
      WHERE user_id = $2 AND currency = $3 AND chain = $4
    `;
    await client.query(query, [amount, userId, currency, chain]);
  }

  // 减少可用余额
  async subtractAvailable(
    client: any,
    userId: number,
    currency: string,
    chain: string,
    amount: string
  ): Promise<void> {
    const query = `
      UPDATE user_assets
      SET available = available - $1, updated_at = NOW()
      WHERE user_id = $2 AND currency = $3 AND chain = $4 AND available >= $1
    `;
    const result = await client.query(query, [amount, userId, currency, chain]);

    if (result.rowCount === 0) {
      throw new Error('Insufficient balance');
    }
  }

  // 冻结余额
  async freezeBalance(
    client: any,
    userId: number,
    currency: string,
    chain: string,
    amount: string
  ): Promise<void> {
    const query = `
      UPDATE user_assets
      SET available = available - $1, frozen = frozen + $1, updated_at = NOW()
      WHERE user_id = $2 AND currency = $3 AND chain = $4 AND available >= $1
    `;
    const result = await client.query(query, [amount, userId, currency, chain]);

    if (result.rowCount === 0) {
      throw new Error('Insufficient balance to freeze');
    }
  }

  // 解冻余额
  async unfreezeBalance(
    client: any,
    userId: number,
    currency: string,
    chain: string,
    amount: string
  ): Promise<void> {
    const query = `
      UPDATE user_assets
      SET available = available + $1, frozen = frozen - $1, updated_at = NOW()
      WHERE user_id = $2 AND currency = $3 AND chain = $4 AND frozen >= $1
    `;
    const result = await client.query(query, [amount, userId, currency, chain]);

    if (result.rowCount === 0) {
      throw new Error('Insufficient frozen balance');
    }
  }

  // 扣除冻结余额
  async deductFrozen(
    client: any,
    userId: number,
    currency: string,
    chain: string,
    amount: string
  ): Promise<void> {
    const query = `
      UPDATE user_assets
      SET frozen = frozen - $1, updated_at = NOW()
      WHERE user_id = $2 AND currency = $3 AND chain = $4 AND frozen >= $1
    `;
    const result = await client.query(query, [amount, userId, currency, chain]);

    if (result.rowCount === 0) {
      throw new Error('Insufficient frozen balance');
    }
  }

  // 记录资产流水
  async logTransaction(
    client: any,
    data: {
      user_id: number;
      currency: string;
      chain: string;
      amount: string;
      balance_before: string;
      balance_after: string;
      tx_type: number;
      ref_id?: number;
      description?: string;
    }
  ): Promise<AssetTransaction> {
    const query = `
      INSERT INTO asset_transactions (
        user_id, currency, chain, amount, balance_before, balance_after,
        tx_type, ref_id, description
      )
      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
      RETURNING *
    `;
    const result = await client.query(query, [
      data.user_id,
      data.currency,
      data.chain,
      data.amount,
      data.balance_before,
      data.balance_after,
      data.tx_type,
      data.ref_id || null,
      data.description || null,
    ]);
    return result.rows[0];
  }

  // 创建充值记录
  async createDeposit(data: {
    user_id: number;
    currency: string;
    chain: string;
    amount: string;
    address: string;
    from_address?: string;
    txid: string;
    required_confirmations: number;
  }): Promise<Deposit> {
    const query = `
      INSERT INTO deposits (
        user_id, currency, chain, amount, address, from_address,
        txid, confirmations, required_confirmations, status
      )
      VALUES ($1, $2, $3, $4, $5, $6, $7, 0, $8, 0)
      RETURNING *
    `;
    const result = await database.query(query, [
      data.user_id,
      data.currency,
      data.chain,
      data.amount,
      data.address,
      data.from_address || null,
      data.txid,
      data.required_confirmations,
    ]);
    return result.rows[0];
  }

  // 更新充值确认数
  async updateDepositConfirmations(
    depositId: number,
    confirmations: number
  ): Promise<Deposit> {
    const query = `
      UPDATE deposits
      SET confirmations = $1, updated_at = NOW()
      WHERE deposit_id = $2
      RETURNING *
    `;
    const result = await database.query(query, [confirmations, depositId]);
    return result.rows[0];
  }

  // 确认充值
  async confirmDeposit(depositId: number): Promise<Deposit> {
    const query = `
      UPDATE deposits
      SET status = 1, confirmed_at = NOW(), updated_at = NOW()
      WHERE deposit_id = $1
      RETURNING *
    `;
    const result = await database.query(query, [depositId]);
    return result.rows[0];
  }

  // 获取用户充值记录
  async getUserDeposits(
    userId: number,
    params: { page?: number; pageSize?: number; currency?: string }
  ): Promise<{ deposits: Deposit[]; total: number }> {
    const page = params.page || 1;
    const pageSize = params.pageSize || 20;
    const offset = (page - 1) * pageSize;

    const conditions = ['user_id = $1'];
    const values: any[] = [userId];
    let paramIndex = 2;

    if (params.currency) {
      conditions.push(`currency = $${paramIndex}`);
      values.push(params.currency);
      paramIndex++;
    }

    const whereClause = conditions.join(' AND ');

    const countQuery = `SELECT COUNT(*) FROM deposits WHERE ${whereClause}`;
    const countResult = await database.query(countQuery, values);
    const total = parseInt(countResult.rows[0].count, 10);

    values.push(pageSize, offset);
    const query = `
      SELECT * FROM deposits
      WHERE ${whereClause}
      ORDER BY created_at DESC
      LIMIT $${paramIndex} OFFSET $${paramIndex + 1}
    `;
    const result = await database.query(query, values);

    return { deposits: result.rows, total };
  }

  // 创建提现申请
  async createWithdrawal(data: {
    user_id: number;
    currency: string;
    chain: string;
    amount: string;
    fee: string;
    address: string;
    address_tag?: string;
  }): Promise<Withdrawal> {
    const actualAmount = new Decimal(data.amount).minus(data.fee).toString();

    const query = `
      INSERT INTO withdrawals (
        user_id, currency, chain, amount, fee, actual_amount,
        address, address_tag, status
      )
      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0)
      RETURNING *
    `;
    const result = await database.query(query, [
      data.user_id,
      data.currency,
      data.chain,
      data.amount,
      data.fee,
      actualAmount,
      data.address,
      data.address_tag || null,
    ]);
    return result.rows[0];
  }

  // 更新提现状态
  async updateWithdrawalStatus(
    withdrawalId: number,
    status: number,
    txid?: string
  ): Promise<Withdrawal> {
    const query = `
      UPDATE withdrawals
      SET status = $1, txid = $2, updated_at = NOW()
      WHERE withdrawal_id = $3
      RETURNING *
    `;
    const result = await database.query(query, [status, txid || null, withdrawalId]);
    return result.rows[0];
  }

  // 完成提现
  async completeWithdrawal(withdrawalId: number, txid: string): Promise<Withdrawal> {
    const query = `
      UPDATE withdrawals
      SET status = 3, txid = $1, complete_time = NOW(), updated_at = NOW()
      WHERE withdrawal_id = $2
      RETURNING *
    `;
    const result = await database.query(query, [txid, withdrawalId]);
    return result.rows[0];
  }

  // 拒绝提现
  async rejectWithdrawal(
    withdrawalId: number,
    auditorId: number,
    reason: string
  ): Promise<Withdrawal> {
    const query = `
      UPDATE withdrawals
      SET status = 4, audit_user_id = $1, reject_reason = $2,
          audit_time = NOW(), updated_at = NOW()
      WHERE withdrawal_id = $3
      RETURNING *
    `;
    const result = await database.query(query, [auditorId, reason, withdrawalId]);
    return result.rows[0];
  }

  // 获取用户提现记录
  async getUserWithdrawals(
    userId: number,
    params: { page?: number; pageSize?: number; currency?: string }
  ): Promise<{ withdrawals: Withdrawal[]; total: number }> {
    const page = params.page || 1;
    const pageSize = params.pageSize || 20;
    const offset = (page - 1) * pageSize;

    const conditions = ['user_id = $1'];
    const values: any[] = [userId];
    let paramIndex = 2;

    if (params.currency) {
      conditions.push(`currency = $${paramIndex}`);
      values.push(params.currency);
      paramIndex++;
    }

    const whereClause = conditions.join(' AND ');

    const countQuery = `SELECT COUNT(*) FROM withdrawals WHERE ${whereClause}`;
    const countResult = await database.query(countQuery, values);
    const total = parseInt(countResult.rows[0].count, 10);

    values.push(pageSize, offset);
    const query = `
      SELECT * FROM withdrawals
      WHERE ${whereClause}
      ORDER BY created_at DESC
      LIMIT $${paramIndex} OFFSET $${paramIndex + 1}
    `;
    const result = await database.query(query, values);

    return { withdrawals: result.rows, total };
  }

  // 获取用户资产流水
  async getUserTransactions(
    userId: number,
    params: {
      page?: number;
      pageSize?: number;
      currency?: string;
      tx_type?: number;
      startDate?: Date;
      endDate?: Date;
    }
  ): Promise<{ transactions: AssetTransaction[]; total: number }> {
    const page = params.page || 1;
    const pageSize = params.pageSize || 50;
    const offset = (page - 1) * pageSize;

    const conditions = ['user_id = $1'];
    const values: any[] = [userId];
    let paramIndex = 2;

    if (params.currency) {
      conditions.push(`currency = $${paramIndex}`);
      values.push(params.currency);
      paramIndex++;
    }

    if (params.tx_type) {
      conditions.push(`tx_type = $${paramIndex}`);
      values.push(params.tx_type);
      paramIndex++;
    }

    if (params.startDate) {
      conditions.push(`created_at >= $${paramIndex}`);
      values.push(params.startDate);
      paramIndex++;
    }

    if (params.endDate) {
      conditions.push(`created_at <= $${paramIndex}`);
      values.push(params.endDate);
      paramIndex++;
    }

    const whereClause = conditions.join(' AND ');

    const countQuery = `SELECT COUNT(*) FROM asset_transactions WHERE ${whereClause}`;
    const countResult = await database.query(countQuery, values);
    const total = parseInt(countResult.rows[0].count, 10);

    values.push(pageSize, offset);
    const query = `
      SELECT * FROM asset_transactions
      WHERE ${whereClause}
      ORDER BY created_at DESC
      LIMIT $${paramIndex} OFFSET $${paramIndex + 1}
    `;
    const result = await database.query(query, values);

    return { transactions: result.rows, total };
  }
}

export default new AssetModel();
