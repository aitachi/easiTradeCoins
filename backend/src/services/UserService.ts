//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

import UserModel, { User, UserSecurity } from '../models/User';
import { AuthUtils } from '../utils/auth';
import redis from '../database/redis';
import logger from '../utils/logger';
import config from '../config';

export class UserService {
  // 注册用户
  async register(data: {
    email: string;
    password: string;
    phone?: string;
    referralCode?: string;
    ip?: string;
  }): Promise<{ user: User; token: string; refreshToken: string }> {
    // 检查邮箱是否已存在
    const existingUser = await UserModel.findByEmail(data.email);
    if (existingUser) {
      throw new Error('Email already registered');
    }

    // 如果提供了手机号,检查是否已存在
    if (data.phone) {
      const existingPhone = await UserModel.findByPhone(data.phone);
      if (existingPhone) {
        throw new Error('Phone number already registered');
      }
    }

    // 处理推荐码
    let referrerId: number | undefined;
    if (data.referralCode) {
      // TODO: 实现推荐码验证
      // const referrer = await this.findByReferralCode(data.referralCode);
      // if (referrer) referrerId = referrer.user_id;
    }

    // 生成密码盐和哈希
    const salt = AuthUtils.generateSalt();
    const passwordHash = await AuthUtils.hashPassword(data.password, salt);

    // 创建用户
    const user = await UserModel.create({
      email: data.email,
      phone: data.phone,
      password_hash: passwordHash,
      salt: salt,
      register_ip: data.ip,
      referrer_id: referrerId,
    });

    // 创建用户安全配置
    await UserModel.createSecurity(user.user_id);

    // 生成Token
    const token = AuthUtils.generateToken({
      userId: user.user_id,
      email: user.email,
    });
    const refreshToken = AuthUtils.generateRefreshToken({
      userId: user.user_id,
    });

    // 存储刷新Token到Redis
    await redis.set(
      `refresh_token:${user.user_id}`,
      refreshToken,
      30 * 24 * 60 * 60 // 30天
    );

    logger.info('User registered', { userId: user.user_id, email: user.email });

    return { user, token, refreshToken };
  }

  // 用户登录
  async login(data: {
    email: string;
    password: string;
    twoFaCode?: string;
    ip?: string;
    userAgent?: string;
    deviceFingerprint?: string;
  }): Promise<{ user: User; token: string; refreshToken: string; require2FA?: boolean }> {
    // 查找用户
    const user = await UserModel.findByEmail(data.email);
    if (!user) {
      throw new Error('Invalid email or password');
    }

    // 检查账户状态
    if (user.status === 2) {
      throw new Error('Account is frozen');
    }
    if (user.status === 3) {
      throw new Error('Account is deactivated');
    }

    // 获取安全配置
    const security = await UserModel.getSecurity(user.user_id);
    if (!security) {
      throw new Error('Security configuration not found');
    }

    // 检查是否被锁定
    if (security.login_locked_until && new Date(security.login_locked_until) > new Date()) {
      const remainingSeconds = Math.ceil(
        (new Date(security.login_locked_until).getTime() - Date.now()) / 1000
      );
      throw new Error(`Account locked. Try again in ${remainingSeconds} seconds`);
    }

    // 验证密码
    const isPasswordValid = await AuthUtils.verifyPassword(
      data.password,
      user.salt,
      user.password_hash
    );

    if (!isPasswordValid) {
      // 增加错误计数
      const errorCount = security.login_password_error_count + 1;
      const updateData: Partial<UserSecurity> = {
        login_password_error_count: errorCount,
      };

      // 如果达到最大尝试次数,锁定账户
      if (errorCount >= config.security.maxLoginAttempts) {
        const lockUntil = new Date(Date.now() + config.security.loginLockDuration * 1000);
        updateData.login_locked_until = lockUntil;
        updateData.login_password_error_count = 0;
      }

      await UserModel.updateSecurity(user.user_id, updateData);

      // 记录登录失败日志
      await UserModel.logLogin({
        user_id: user.user_id,
        login_ip: data.ip || '',
        user_agent: data.userAgent,
        device_fingerprint: data.deviceFingerprint,
        login_status: 2,
        fail_reason: 'Invalid password',
      });

      throw new Error('Invalid email or password');
    }

    // 检查是否需要2FA
    if (security.is_2fa_enabled) {
      if (!data.twoFaCode) {
        return {
          user,
          token: '',
          refreshToken: '',
          require2FA: true,
        };
      }

      // 验证2FA代码
      const is2FAValid = AuthUtils.verify2FACode(security.google_2fa_secret!, data.twoFaCode);
      if (!is2FAValid) {
        await UserModel.logLogin({
          user_id: user.user_id,
          login_ip: data.ip || '',
          user_agent: data.userAgent,
          device_fingerprint: data.deviceFingerprint,
          login_status: 2,
          fail_reason: 'Invalid 2FA code',
        });
        throw new Error('Invalid 2FA code');
      }
    }

    // 重置错误计数
    if (security.login_password_error_count > 0) {
      await UserModel.updateSecurity(user.user_id, {
        login_password_error_count: 0,
        login_locked_until: null,
      });
    }

    // 更新最后登录时间
    await UserModel.updateLastLogin(user.user_id, data.ip || '');

    // 记录登录成功日志
    await UserModel.logLogin({
      user_id: user.user_id,
      login_ip: data.ip || '',
      user_agent: data.userAgent,
      device_fingerprint: data.deviceFingerprint,
      login_status: 1,
    });

    // 生成Token
    const token = AuthUtils.generateToken({
      userId: user.user_id,
      email: user.email,
      kycLevel: user.kyc_level,
      vipLevel: user.vip_level,
    });
    const refreshToken = AuthUtils.generateRefreshToken({
      userId: user.user_id,
    });

    // 存储刷新Token到Redis
    await redis.set(
      `refresh_token:${user.user_id}`,
      refreshToken,
      30 * 24 * 60 * 60 // 30天
    );

    logger.info('User logged in', { userId: user.user_id, email: user.email });

    return { user, token, refreshToken };
  }

  // 刷新Token
  async refreshToken(refreshToken: string): Promise<{ token: string; refreshToken: string }> {
    const payload = AuthUtils.verifyRefreshToken(refreshToken);
    if (!payload) {
      throw new Error('Invalid refresh token');
    }

    // 检查Redis中的刷新Token
    const storedToken = await redis.get(`refresh_token:${payload.userId}`);
    if (storedToken !== refreshToken) {
      throw new Error('Refresh token not found or expired');
    }

    // 获取用户信息
    const user = await UserModel.findById(payload.userId);
    if (!user) {
      throw new Error('User not found');
    }

    // 生成新Token
    const newToken = AuthUtils.generateToken({
      userId: user.user_id,
      email: user.email,
      kycLevel: user.kyc_level,
      vipLevel: user.vip_level,
    });
    const newRefreshToken = AuthUtils.generateRefreshToken({
      userId: user.user_id,
    });

    // 更新Redis中的刷新Token
    await redis.set(
      `refresh_token:${user.user_id}`,
      newRefreshToken,
      30 * 24 * 60 * 60
    );

    return { token: newToken, refreshToken: newRefreshToken };
  }

  // 登出
  async logout(userId: number): Promise<void> {
    await redis.del(`refresh_token:${userId}`);
    logger.info('User logged out', { userId });
  }

  // 启用2FA
  async enable2FA(userId: number): Promise<{ secret: string; qrCode: string }> {
    const user = await UserModel.findById(userId);
    if (!user) {
      throw new Error('User not found');
    }

    // 生成2FA密钥
    const { secret, qrCode: otpauthUrl } = AuthUtils.generate2FASecret(user.email);

    // 生成二维码
    const qrCode = await AuthUtils.generate2FAQRCode(secret, user.email);

    // 临时存储密钥到Redis,待验证后正式启用
    await redis.set(`2fa_temp:${userId}`, secret, 300); // 5分钟过期

    return { secret, qrCode };
  }

  // 验证并激活2FA
  async verify2FA(userId: number, code: string): Promise<void> {
    const tempSecret = await redis.get(`2fa_temp:${userId}`);
    if (!tempSecret) {
      throw new Error('2FA setup expired, please start again');
    }

    const isValid = AuthUtils.verify2FACode(tempSecret, code);
    if (!isValid) {
      throw new Error('Invalid 2FA code');
    }

    // 启用2FA
    await UserModel.updateSecurity(userId, {
      google_2fa_secret: tempSecret,
      is_2fa_enabled: true,
    });

    // 删除临时密钥
    await redis.del(`2fa_temp:${userId}`);

    logger.info('2FA enabled', { userId });
  }

  // 禁用2FA
  async disable2FA(userId: number, code: string): Promise<void> {
    const security = await UserModel.getSecurity(userId);
    if (!security || !security.is_2fa_enabled) {
      throw new Error('2FA not enabled');
    }

    const isValid = AuthUtils.verify2FACode(security.google_2fa_secret!, code);
    if (!isValid) {
      throw new Error('Invalid 2FA code');
    }

    await UserModel.updateSecurity(userId, {
      is_2fa_enabled: false,
    });

    logger.info('2FA disabled', { userId });
  }

  // 获取用户信息
  async getUserProfile(userId: number): Promise<User> {
    const user = await UserModel.findById(userId);
    if (!user) {
      throw new Error('User not found');
    }
    return user;
  }

  // 更新用户资料
  async updateProfile(
    userId: number,
    data: { nickname?: string; avatar_url?: string }
  ): Promise<User> {
    return await UserModel.update(userId, data);
  }

  // 修改密码
  async changePassword(
    userId: number,
    oldPassword: string,
    newPassword: string
  ): Promise<void> {
    const user = await UserModel.findById(userId);
    if (!user) {
      throw new Error('User not found');
    }

    // 验证旧密码
    const isValid = await AuthUtils.verifyPassword(oldPassword, user.salt, user.password_hash);
    if (!isValid) {
      throw new Error('Invalid old password');
    }

    // 生成新密码哈希
    const newSalt = AuthUtils.generateSalt();
    const newHash = await AuthUtils.hashPassword(newPassword, newSalt);

    await UserModel.update(userId, {
      password_hash: newHash,
      salt: newSalt,
    });

    logger.info('Password changed', { userId });
  }

  // 提交KYC申请
  async submitKYC(data: {
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
  }): Promise<void> {
    // 检查是否已有待审核的KYC
    const existingKYC = await UserModel.getKYC(data.user_id);
    if (existingKYC && existingKYC.audit_status === 0) {
      throw new Error('You already have a pending KYC application');
    }

    await UserModel.createKYC(data);
    logger.info('KYC submitted', { userId: data.user_id });
  }
}

export default new UserService();
