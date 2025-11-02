//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

import dotenv from 'dotenv';
import path from 'path';

// 加载环境变量
dotenv.config();

interface Config {
  env: string;
  port: number;
  apiPrefix: string;

  // 数据库配置
  database: {
    host: string;
    port: number;
    database: string;
    user: string;
    password: string;
    max: number;
    idleTimeoutMillis: number;
  };

  // Redis配置
  redis: {
    host: string;
    port: number;
    password: string;
    db: number;
  };

  // JWT配置
  jwt: {
    secret: string;
    expiresIn: string;
    refreshSecret: string;
    refreshExpiresIn: string;
  };

  // 安全配置
  security: {
    bcryptRounds: number;
    maxLoginAttempts: number;
    loginLockDuration: number;
    sessionTimeout: number;
    apiRateLimitPublic: number;
    apiRateLimitPrivate: number;
  };

  // 区块链配置
  blockchain: {
    ethereum: {
      mainnetRpc: string;
      sepoliaRpc: string;
      privateKey: string;
    };
    solana: {
      mainnetRpc: string;
      devnetRpc: string;
      privateKey: string;
    };
  };

  // 交易配置
  trading: {
    orderExpirySeconds: number;
    maxOpenOrdersPerUser: number;
    minOrderAmountUSD: number;
  };

  // 提现配置
  withdrawal: {
    manualReviewThreshold: number;
    delayHours: number;
    hotWalletPercentage: number;
    coldWalletPercentage: number;
  };

  // 风控配置
  riskControl: {
    enabled: boolean;
    amlEnabled: boolean;
    maxDailyWithdrawUSD: number;
  };

  // WebSocket配置
  websocket: {
    port: number;
    heartbeatInterval: number;
  };

  // 日志配置
  logging: {
    level: string;
    filePath: string;
  };

  // CORS配置
  cors: {
    origin: string;
  };

  // AI配置
  ai: {
    volcanoApiUrl: string;
    volcanoApiKey: string;
    volcanoModel: string;
  };

  // 邮件配置
  email: {
    host: string;
    port: number;
    secure: boolean;
    user: string;
    password: string;
    from: string;
  };
}

const config: Config = {
  env: process.env.NODE_ENV || 'development',
  port: parseInt(process.env.PORT || '3000', 10),
  apiPrefix: process.env.API_PREFIX || '/api/v1',

  database: {
    host: process.env.DB_HOST || 'localhost',
    port: parseInt(process.env.DB_PORT || '5432', 10),
    database: process.env.DB_NAME || 'easitradecoins',
    user: process.env.DB_USER || 'postgres',
    password: process.env.DB_PASSWORD || '',
    max: 20,
    idleTimeoutMillis: 30000,
  },

  redis: {
    host: process.env.REDIS_HOST || 'localhost',
    port: parseInt(process.env.REDIS_PORT || '6379', 10),
    password: process.env.REDIS_PASSWORD || '',
    db: parseInt(process.env.REDIS_DB || '0', 10),
  },

  jwt: {
    secret: process.env.JWT_SECRET || 'your-secret-key',
    expiresIn: process.env.JWT_EXPIRES_IN || '7d',
    refreshSecret: process.env.JWT_REFRESH_SECRET || 'your-refresh-secret',
    refreshExpiresIn: process.env.JWT_REFRESH_EXPIRES_IN || '30d',
  },

  security: {
    bcryptRounds: parseInt(process.env.BCRYPT_ROUNDS || '10', 10),
    maxLoginAttempts: parseInt(process.env.MAX_LOGIN_ATTEMPTS || '5', 10),
    loginLockDuration: parseInt(process.env.LOGIN_LOCK_DURATION || '1800', 10),
    sessionTimeout: parseInt(process.env.SESSION_TIMEOUT || '7200', 10),
    apiRateLimitPublic: parseInt(process.env.API_RATE_LIMIT_PUBLIC || '100', 10),
    apiRateLimitPrivate: parseInt(process.env.API_RATE_LIMIT_PRIVATE || '50', 10),
  },

  blockchain: {
    ethereum: {
      mainnetRpc: process.env.ETH_MAINNET_RPC || '',
      sepoliaRpc: process.env.ETH_SEPOLIA_RPC || '',
      privateKey: process.env.ETH_PRIVATE_KEY || '',
    },
    solana: {
      mainnetRpc: process.env.SOLANA_MAINNET_RPC || '',
      devnetRpc: process.env.SOLANA_DEVNET_RPC || '',
      privateKey: process.env.SOLANA_PRIVATE_KEY || '',
    },
  },

  trading: {
    orderExpirySeconds: parseInt(process.env.ORDER_EXPIRY_SECONDS || '86400', 10),
    maxOpenOrdersPerUser: parseInt(process.env.MAX_OPEN_ORDERS_PER_USER || '100', 10),
    minOrderAmountUSD: parseFloat(process.env.MIN_ORDER_AMOUNT_USD || '10'),
  },

  withdrawal: {
    manualReviewThreshold: parseFloat(process.env.WITHDRAW_MANUAL_REVIEW_THRESHOLD || '10000'),
    delayHours: parseInt(process.env.WITHDRAW_DELAY_HOURS || '24', 10),
    hotWalletPercentage: parseInt(process.env.HOT_WALLET_PERCENTAGE || '10', 10),
    coldWalletPercentage: parseInt(process.env.COLD_WALLET_PERCENTAGE || '60', 10),
  },

  riskControl: {
    enabled: process.env.RISK_CHECK_ENABLED === 'true',
    amlEnabled: process.env.AML_ENABLED === 'true',
    maxDailyWithdrawUSD: parseFloat(process.env.MAX_DAILY_WITHDRAW_USD || '100000'),
  },

  websocket: {
    port: parseInt(process.env.WS_PORT || '3001', 10),
    heartbeatInterval: parseInt(process.env.WS_HEARTBEAT_INTERVAL || '30000', 10),
  },

  logging: {
    level: process.env.LOG_LEVEL || 'info',
    filePath: process.env.LOG_FILE_PATH || './logs',
  },

  cors: {
    origin: process.env.FRONTEND_URL || 'http://localhost:3000',
  },

  ai: {
    volcanoApiUrl: process.env.VOLCANO_API_URL || '',
    volcanoApiKey: process.env.VOLCANO_API_KEY || '',
    volcanoModel: process.env.VOLCANO_MODEL || '',
  },

  email: {
    host: process.env.SMTP_HOST || '',
    port: parseInt(process.env.SMTP_PORT || '587', 10),
    secure: process.env.SMTP_SECURE === 'true',
    user: process.env.SMTP_USER || '',
    password: process.env.SMTP_PASSWORD || '',
    from: process.env.SMTP_FROM || '',
  },
};

export default config;
