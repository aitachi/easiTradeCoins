import database from '../database';
import { QueryResult } from 'pg';

export interface User {
  user_id: number;
  email: string;
  phone?: string;
  password_hash: string;
  salt: string;
  kyc_level: number;
  status: number;
  vip_level: number;
  register_ip?: string;
  register_time: Date;
  last_login_time?: Date;
  last_login_ip?: string;
  referrer_id?: number;
  avatar_url?: string;
  nickname?: string;
  created_at: Date;
  updated_at: Date;
}

export interface UserSecurity {
  user_id: number;
  google_2fa_secret?: string;
  is_2fa_enabled: boolean;
  withdrawal_whitelist?: any;
  api_key_hash?: string;
  api_secret_hash?: string;
  api_permissions?: any;
  login_password_error_count: number;
  login_locked_until?: Date;
  device_fingerprints?: any[];
  created_at: Date;
  updated_at: Date;
}

export interface UserKYC {
  kyc_id: number;
  user_id: number;
  real_name?: string;
  id_number?: string;
  id_type?: number;
  id_front_url?: string;
  id_back_url?: string;
  face_url?: string;
  country_code?: string;
  address?: string;
  date_of_birth?: Date;
  submit_time: Date;
  audit_time?: Date;
  audit_status: number;
  audit_remark?: string;
  auditor_id?: number;
  created_at: Date;
  updated_at: Date;
}

class UserModel {
  // 创建用户
  async create(data: {
    email: string;
    phone?: string;
    password_hash: string;
    salt: string;
    register_ip?: string;
    referrer_id?: number;
  }): Promise<User> {
    const query = `
      INSERT INTO users (email, phone, password_hash, salt, register_ip, referrer_id)
      VALUES ($1, $2, $3, $4, $5, $6)
      RETURNING *
    `;
    const result = await database.query(query, [
      data.email,
      data.phone || null,
      data.password_hash,
      data.salt,
      data.register_ip || null,
      data.referrer_id || null,
    ]);
    return result.rows[0];
  }

  // 根据ID查询用户
  async findById(userId: number): Promise<User | null> {
    const query = 'SELECT * FROM users WHERE user_id = $1';
    const result = await database.query(query, [userId]);
    return result.rows[0] || null;
  }

  // 根据邮箱查询用户
  async findByEmail(email: string): Promise<User | null> {
    const query = 'SELECT * FROM users WHERE email = $1';
    const result = await database.query(query, [email]);
    return result.rows[0] || null;
  }

  // 根据手机号查询用户
  async findByPhone(phone: string): Promise<User | null> {
    const query = 'SELECT * FROM users WHERE phone = $1';
    const result = await database.query(query, [phone]);
    return result.rows[0] || null;
  }

  // 更新用户信息
  async update(userId: number, data: Partial<User>): Promise<User> {
    const fields: string[] = [];
    const values: any[] = [];
    let paramIndex = 1;

    Object.entries(data).forEach(([key, value]) => {
      if (value !== undefined && key !== 'user_id' && key !== 'created_at') {
        fields.push(`${key} = $${paramIndex}`);
        values.push(value);
        paramIndex++;
      }
    });

    if (fields.length === 0) {
      return (await this.findById(userId))!;
    }

    values.push(userId);
    const query = `
      UPDATE users
      SET ${fields.join(', ')}, updated_at = NOW()
      WHERE user_id = $${paramIndex}
      RETURNING *
    `;
    const result = await database.query(query, values);
    return result.rows[0];
  }

  // 更新最后登录时间
  async updateLastLogin(userId: number, ip: string): Promise<void> {
    const query = `
      UPDATE users
      SET last_login_time = NOW(), last_login_ip = $1
      WHERE user_id = $2
    `;
    await database.query(query, [ip, userId]);
  }

  // 创建用户安全配置
  async createSecurity(userId: number): Promise<UserSecurity> {
    const query = `
      INSERT INTO user_security (user_id)
      VALUES ($1)
      RETURNING *
    `;
    const result = await database.query(query, [userId]);
    return result.rows[0];
  }

  // 获取用户安全配置
  async getSecurity(userId: number): Promise<UserSecurity | null> {
    const query = 'SELECT * FROM user_security WHERE user_id = $1';
    const result = await database.query(query, [userId]);
    return result.rows[0] || null;
  }

  // 更新用户安全配置
  async updateSecurity(userId: number, data: Partial<UserSecurity>): Promise<UserSecurity> {
    const fields: string[] = [];
    const values: any[] = [];
    let paramIndex = 1;

    Object.entries(data).forEach(([key, value]) => {
      if (value !== undefined && key !== 'user_id' && key !== 'created_at') {
        fields.push(`${key} = $${paramIndex}`);
        values.push(value);
        paramIndex++;
      }
    });

    values.push(userId);
    const query = `
      UPDATE user_security
      SET ${fields.join(', ')}, updated_at = NOW()
      WHERE user_id = $${paramIndex}
      RETURNING *
    `;
    const result = await database.query(query, values);
    return result.rows[0];
  }

  // 创建KYC记录
  async createKYC(data: {
    user_id: number;
    real_name: string;
    id_number: string;
    id_type: number;
    id_front_url?: string;
    id_back_url?: string;
    face_url?: string;
    country_code: string;
    address?: string;
    date_of_birth?: Date;
  }): Promise<UserKYC> {
    const query = `
      INSERT INTO user_kyc (
        user_id, real_name, id_number, id_type, id_front_url, id_back_url,
        face_url, country_code, address, date_of_birth, audit_status
      )
      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 0)
      RETURNING *
    `;
    const result = await database.query(query, [
      data.user_id,
      data.real_name,
      data.id_number,
      data.id_type,
      data.id_front_url || null,
      data.id_back_url || null,
      data.face_url || null,
      data.country_code,
      data.address || null,
      data.date_of_birth || null,
    ]);
    return result.rows[0];
  }

  // 获取用户KYC信息
  async getKYC(userId: number): Promise<UserKYC | null> {
    const query = 'SELECT * FROM user_kyc WHERE user_id = $1 ORDER BY submit_time DESC LIMIT 1';
    const result = await database.query(query, [userId]);
    return result.rows[0] || null;
  }

  // 更新KYC审核状态
  async updateKYCAudit(
    kycId: number,
    auditorId: number,
    status: number,
    remark?: string
  ): Promise<UserKYC> {
    const query = `
      UPDATE user_kyc
      SET audit_status = $1, auditor_id = $2, audit_remark = $3, audit_time = NOW()
      WHERE kyc_id = $4
      RETURNING *
    `;
    const result = await database.query(query, [status, auditorId, remark || null, kycId]);

    // 如果审核通过,更新用户KYC等级
    if (status === 1) {
      const kyc = result.rows[0];
      await this.update(kyc.user_id, { kyc_level: 1 });
    }

    return result.rows[0];
  }

  // 记录登录日志
  async logLogin(data: {
    user_id: number;
    login_ip: string;
    login_location?: string;
    device_type?: string;
    device_fingerprint?: string;
    user_agent?: string;
    login_status: number;
    fail_reason?: string;
  }): Promise<void> {
    const query = `
      INSERT INTO user_login_logs (
        user_id, login_ip, login_location, device_type, device_fingerprint,
        user_agent, login_status, fail_reason
      )
      VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `;
    await database.query(query, [
      data.user_id,
      data.login_ip,
      data.login_location || null,
      data.device_type || null,
      data.device_fingerprint || null,
      data.user_agent || null,
      data.login_status,
      data.fail_reason || null,
    ]);
  }

  // 获取用户列表
  async list(params: {
    page?: number;
    pageSize?: number;
    status?: number;
    kycLevel?: number;
  }): Promise<{ users: User[]; total: number }> {
    const page = params.page || 1;
    const pageSize = params.pageSize || 20;
    const offset = (page - 1) * pageSize;

    const conditions: string[] = [];
    const values: any[] = [];
    let paramIndex = 1;

    if (params.status !== undefined) {
      conditions.push(`status = $${paramIndex}`);
      values.push(params.status);
      paramIndex++;
    }

    if (params.kycLevel !== undefined) {
      conditions.push(`kyc_level = $${paramIndex}`);
      values.push(params.kycLevel);
      paramIndex++;
    }

    const whereClause = conditions.length > 0 ? `WHERE ${conditions.join(' AND ')}` : '';

    const countQuery = `SELECT COUNT(*) FROM users ${whereClause}`;
    const countResult = await database.query(countQuery, values);
    const total = parseInt(countResult.rows[0].count, 10);

    values.push(pageSize, offset);
    const query = `
      SELECT * FROM users
      ${whereClause}
      ORDER BY created_at DESC
      LIMIT $${paramIndex} OFFSET $${paramIndex + 1}
    `;
    const result = await database.query(query, values);

    return { users: result.rows, total };
  }
}

export default new UserModel();
