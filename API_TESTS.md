# EasiTradeCoins API Test Collection

This file contains example API requests for testing the EasiTradeCoins platform.

## Prerequisites

```bash
# Set these variables
export API_URL="http://localhost:8080"
export TOKEN="your-jwt-token-here"
```

## 1. Health Check

```bash
curl -X GET $API_URL/health
```

Expected Response:
```json
{
  "status": "ok"
}
```

## 2. User Registration

```bash
curl -X POST $API_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "trader@example.com",
    "password": "SecurePass123!",
    "phone": "+1234567890"
  }'
```

Expected Response:
```json
{
  "user": {
    "id": 1,
    "email": "trader@example.com",
    "kyc_level": 0,
    "status": 1
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## 3. User Login

```bash
curl -X POST $API_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "trader@example.com",
    "password": "SecurePass123!"
  }'
```

Save the token from the response:
```bash
export TOKEN="<token-from-response>"
```

## 4. Get Account Balance

```bash
curl -X GET $API_URL/api/v1/account/balance \
  -H "Authorization: Bearer $TOKEN"
```

Expected Response:
```json
[
  {
    "id": 1,
    "user_id": 1,
    "currency": "BTC",
    "chain": "ERC20",
    "available": "0.00000000",
    "frozen": "0.00000000"
  },
  {
    "id": 2,
    "user_id": 1,
    "currency": "ETH",
    "chain": "ERC20",
    "available": "0.00000000",
    "frozen": "0.00000000"
  },
  {
    "id": 3,
    "user_id": 1,
    "currency": "USDT",
    "chain": "ERC20",
    "available": "10000.00000000",
    "frozen": "0.00000000"
  }
]
```

## 5. Create Limit Buy Order

```bash
curl -X POST $API_URL/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTC_USDT",
    "side": "buy",
    "type": "limit",
    "price": "45000.00",
    "quantity": "0.1",
    "timeInForce": "GTC"
  }'
```

Expected Response:
```json
{
  "order": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": 1,
    "symbol": "BTC_USDT",
    "side": "buy",
    "type": "limit",
    "price": "45000.00",
    "quantity": "0.1",
    "filled_qty": "0",
    "status": "pending",
    "create_time": "2025-11-01T12:00:00Z"
  },
  "trades": []
}
```

Save the order ID:
```bash
export ORDER_ID="<order-id-from-response>"
```

## 6. Create Limit Sell Order

```bash
curl -X POST $API_URL/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTC_USDT",
    "side": "sell",
    "type": "limit",
    "price": "44000.00",
    "quantity": "0.05",
    "timeInForce": "GTC"
  }'
```

## 7. Create Market Order

```bash
curl -X POST $API_URL/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "BTC_USDT",
    "side": "buy",
    "type": "market",
    "quantity": "0.01"
  }'
```

## 8. Get Order Details

```bash
curl -X GET $API_URL/api/v1/order/$ORDER_ID \
  -H "Authorization: Bearer $TOKEN"
```

## 9. Get Open Orders

```bash
# All open orders
curl -X GET $API_URL/api/v1/order/open \
  -H "Authorization: Bearer $TOKEN"

# Open orders for specific symbol
curl -X GET "$API_URL/api/v1/order/open?symbol=BTC_USDT" \
  -H "Authorization: Bearer $TOKEN"
```

## 10. Get Order History

```bash
# Get all order history
curl -X GET "$API_URL/api/v1/order/history?limit=50&offset=0" \
  -H "Authorization: Bearer $TOKEN"

# Get order history for specific symbol
curl -X GET "$API_URL/api/v1/order/history?symbol=BTC_USDT&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

## 11. Cancel Order

```bash
curl -X DELETE $API_URL/api/v1/order/$ORDER_ID \
  -H "Authorization: Bearer $TOKEN"
```

Expected Response:
```json
{
  "message": "Order cancelled successfully"
}
```

## 12. Get Order Book Depth

```bash
# Default depth (20 levels)
curl -X GET $API_URL/api/v1/market/depth/BTC_USDT

# Custom depth (50 levels)
curl -X GET "$API_URL/api/v1/market/depth/BTC_USDT?depth=50"
```

Expected Response:
```json
{
  "symbol": "BTC_USDT",
  "bids": [
    {
      "price": "44999.00",
      "volume": "1.5",
      "count": 3
    },
    {
      "price": "44998.00",
      "volume": "2.3",
      "count": 5
    }
  ],
  "asks": [
    {
      "price": "45001.00",
      "volume": "1.2",
      "count": 2
    },
    {
      "price": "45002.00",
      "volume": "3.1",
      "count": 4
    }
  ]
}
```

## 13. Get Recent Trades

```bash
# Default limit (50 trades)
curl -X GET $API_URL/api/v1/market/trades/BTC_USDT

# Custom limit (100 trades)
curl -X GET "$API_URL/api/v1/market/trades/BTC_USDT?limit=100"
```

Expected Response:
```json
[
  {
    "id": "trade-id-1",
    "symbol": "BTC_USDT",
    "price": "45000.50",
    "quantity": "0.15",
    "amount": "6750.075",
    "trade_time": "2025-11-01T12:05:30Z"
  },
  {
    "id": "trade-id-2",
    "symbol": "BTC_USDT",
    "price": "45000.00",
    "quantity": "0.08",
    "amount": "3600.00",
    "trade_time": "2025-11-01T12:05:25Z"
  }
]
```

## WebSocket Connection

### Connect to WebSocket

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('WebSocket connected');

  // Subscribe to channels
  ws.send(JSON.stringify({
    method: 'SUBSCRIBE',
    params: [
      'btc_usdt@ticker',
      'btc_usdt@depth',
      'btc_usdt@trade'
    ],
    id: 1
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);

  // Handle different message types
  switch(data.type) {
    case 'subscribed':
      console.log('Subscribed to:', data.data);
      break;
    case 'update':
      console.log('Update:', data.channel, data.data);
      break;
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket disconnected');
};
```

### Subscribe to Additional Channels

```javascript
ws.send(JSON.stringify({
  method: 'SUBSCRIBE',
  params: ['eth_usdt@trade'],
  id: 2
}));
```

### Unsubscribe from Channels

```javascript
ws.send(JSON.stringify({
  method: 'UNSUBSCRIBE',
  params: ['btc_usdt@ticker'],
  id: 3
}));
```

## Advanced Testing Scenarios

### Scenario 1: Complete Trade Flow

```bash
# 1. Register two users
# User A (Buyer)
curl -X POST $API_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"buyer@test.com","password":"Test123456"}'

export TOKEN_BUYER="<token-from-response>"

# User B (Seller)
curl -X POST $API_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"seller@test.com","password":"Test123456"}'

export TOKEN_SELLER="<token-from-response>"

# 2. User A places buy order
curl -X POST $API_URL/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN_BUYER" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTC_USDT","side":"buy","type":"limit","price":"45000","quantity":"0.1"}'

# 3. User B places matching sell order (should trigger trade)
curl -X POST $API_URL/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN_SELLER" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTC_USDT","side":"sell","type":"limit","price":"45000","quantity":"0.1"}'

# 4. Check balances
curl -X GET $API_URL/api/v1/account/balance \
  -H "Authorization: Bearer $TOKEN_BUYER"

curl -X GET $API_URL/api/v1/account/balance \
  -H "Authorization: Bearer $TOKEN_SELLER"
```

### Scenario 2: Order Book Management

```bash
# 1. Place multiple buy orders at different prices
for price in 44000 44100 44200 44300 44400; do
  curl -X POST $API_URL/api/v1/order/create \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"symbol\":\"BTC_USDT\",\"side\":\"buy\",\"type\":\"limit\",\"price\":\"$price\",\"quantity\":\"0.1\"}"
done

# 2. Place multiple sell orders at different prices
for price in 45100 45200 45300 45400 45500; do
  curl -X POST $API_URL/api/v1/order/create \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"symbol\":\"BTC_USDT\",\"side\":\"sell\",\"type\":\"limit\",\"price\":\"$price\",\"quantity\":\"0.1\"}"
done

# 3. Check order book depth
curl -X GET "$API_URL/api/v1/market/depth/BTC_USDT?depth=10"
```

### Scenario 3: Market Order Execution

```bash
# 1. Place limit sell orders
curl -X POST $API_URL/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN_SELLER" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTC_USDT","side":"sell","type":"limit","price":"45000","quantity":"0.5"}'

# 2. Execute market buy order (should match against limit sell)
curl -X POST $API_URL/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN_BUYER" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTC_USDT","side":"buy","type":"market","quantity":"0.2"}'
```

## Error Handling Examples

### 1. Invalid Credentials

```bash
curl -X POST $API_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"wrong@test.com","password":"wrongpass"}'
```

Expected: `401 Unauthorized`

### 2. Insufficient Balance

```bash
curl -X POST $API_URL/api/v1/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTC_USDT","side":"buy","type":"limit","price":"45000","quantity":"1000"}'
```

Expected: `400 Bad Request` with "insufficient balance"

### 3. Invalid Token

```bash
curl -X GET $API_URL/api/v1/account/balance \
  -H "Authorization: Bearer invalid-token"
```

Expected: `401 Unauthorized`

### 4. Cancel Already Filled Order

```bash
# Try to cancel an order that's already filled
curl -X DELETE $API_URL/api/v1/order/<filled-order-id> \
  -H "Authorization: Bearer $TOKEN"
```

Expected: `400 Bad Request` with "order cannot be cancelled"

## Performance Testing

### Load Test with Apache Bench

```bash
# Register endpoint
ab -n 1000 -c 10 -T "application/json" \
  -p register.json \
  $API_URL/api/v1/auth/register

# Market data endpoint (public)
ab -n 10000 -c 100 \
  $API_URL/api/v1/market/depth/BTC_USDT
```

### Create test data file

```bash
# register.json
cat > register.json << EOF
{
  "email": "test@example.com",
  "password": "Test123456"
}
EOF
```

## Monitoring & Debugging

### Check Service Health

```bash
# Backend health
curl $API_URL/health

# Database connection
docker-compose exec postgres psql -U postgres -d easitradecoins -c "SELECT 1"

# Redis connection
docker-compose exec redis redis-cli ping
```

### View Logs

```bash
# Backend logs
docker-compose logs -f backend

# All services
docker-compose logs -f

# Last 100 lines
docker-compose logs --tail=100 backend
```

## Notes

- Replace `$API_URL` with your actual API URL
- Replace `$TOKEN` with actual JWT token from login response
- All timestamps are in UTC
- Decimal places for prices and quantities follow trading pair precision settings
- WebSocket connections automatically reconnect on disconnect

## Support

For issues or questions:
- GitHub: https://github.com/yourusername/EasiTradeCoins/issues
- Email: support@easitradecoins.com
