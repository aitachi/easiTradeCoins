import bcrypt from 'bcrypt';
import jwt from 'jsonwebtoken';
import crypto from 'crypto';
import config from '../config';
import speakeasy from 'speakeasy';
import qrcode from 'qrcode';

export class AuthUtils {
  // 生成随机盐
  static generateSalt(): string {
    return crypto.randomBytes(32).toString('hex');
  }

  // 哈希密码
  static async hashPassword(password: string, salt: string): Promise<string> {
    return await bcrypt.hash(password + salt, config.security.bcryptRounds);
  }

  // 验证密码
  static async verifyPassword(
    password: string,
    salt: string,
    hash: string
  ): Promise<boolean> {
    return await bcrypt.compare(password + salt, hash);
  }

  // 生成JWT Token
  static generateToken(payload: any, expiresIn?: string): string {
    return jwt.sign(payload, config.jwt.secret, {
      expiresIn: expiresIn || config.jwt.expiresIn,
    });
  }

  // 验证JWT Token
  static verifyToken(token: string): any {
    try {
      return jwt.verify(token, config.jwt.secret);
    } catch (error) {
      return null;
    }
  }

  // 生成刷新Token
  static generateRefreshToken(payload: any): string {
    return jwt.sign(payload, config.jwt.refreshSecret, {
      expiresIn: config.jwt.refreshExpiresIn,
    });
  }

  // 验证刷新Token
  static verifyRefreshToken(token: string): any {
    try {
      return jwt.verify(token, config.jwt.refreshSecret);
    } catch (error) {
      return null;
    }
  }

  // 生成2FA密钥
  static generate2FASecret(email: string): { secret: string; qrCode: string } {
    const secret = speakeasy.generateSecret({
      name: `EasiTradeCoins (${email})`,
      issuer: 'EasiTradeCoins',
      length: 32,
    });

    return {
      secret: secret.base32,
      qrCode: secret.otpauth_url || '',
    };
  }

  // 生成2FA二维码
  static async generate2FAQRCode(secret: string, email: string): Promise<string> {
    const otpauth = speakeasy.otpauthURL({
      secret: secret,
      label: email,
      issuer: 'EasiTradeCoins',
      encoding: 'base32',
    });

    return await qrcode.toDataURL(otpauth);
  }

  // 验证2FA代码
  static verify2FACode(secret: string, code: string): boolean {
    return speakeasy.totp.verify({
      secret: secret,
      encoding: 'base32',
      token: code,
      window: 2, // 允许前后2个时间窗口
    });
  }

  // 生成API Key
  static generateApiKey(): string {
    return `ETC_${crypto.randomBytes(24).toString('hex')}`;
  }

  // 生成API Secret
  static generateApiSecret(): string {
    return crypto.randomBytes(32).toString('hex');
  }

  // 哈希API Secret
  static async hashApiSecret(secret: string): Promise<string> {
    return await bcrypt.hash(secret, config.security.bcryptRounds);
  }

  // 生成设备指纹
  static generateDeviceFingerprint(data: {
    userAgent: string;
    ip: string;
    acceptLanguage?: string;
  }): string {
    const fingerprint = `${data.userAgent}|${data.ip}|${data.acceptLanguage || ''}`;
    return crypto.createHash('sha256').update(fingerprint).digest('hex');
  }

  // 生成验证码
  static generateVerificationCode(length: number = 6): string {
    const digits = '0123456789';
    let code = '';
    for (let i = 0; i < length; i++) {
      code += digits[Math.floor(Math.random() * digits.length)];
    }
    return code;
  }

  // 生成随机Token
  static generateRandomToken(length: number = 32): string {
    return crypto.randomBytes(length).toString('hex');
  }
}
