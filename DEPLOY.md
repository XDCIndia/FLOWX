# FlowX Deployment Guide

## Quick Deploy to Render

### Step 1: Fork the Repository
1. Go to https://github.com/XDCIndia/FLOWX
2. Click **"Fork"** to create your own copy

### Step 2: Create Render Account
1. Go to https://render.com
2. Sign up with GitHub
3. Click **"New"** → **"Blueprint"**
4. Select your forked repository
5. Render will auto-detect `render.yaml` and create all services

### Step 3: Set Environment Variables
1. Go to **flowx-api** service
2. Click **"Environment"** tab
3. Add these variables:
   ```
   XDC_TREASURY_SECRET_KEY=your_treasury_private_key
   ```

### Step 4: Deploy
1. Render will automatically deploy all services
2. Wait 5-10 minutes for first deployment
3. Your app will be live at:
   - Frontend: `https://flowx-web.onrender.com`
   - API: `https://flowx-api.onrender.com`

## Manual Deploy (Alternative)

### Backend API
```bash
# Build and run locally
docker build -t flowx-api .
docker run -p 3000:3000 \
  -e DATABASE_URL=your_db_url \
  -e REDIS_URL=your_redis_url \
  flowx-api
```

### Frontend
```bash
cd apps/web
npm install
npm run build
npm start
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `REDIS_URL` | Redis connection string | Yes |
| `PORT` | Server port (default: 3000) | No |
| `XDC_TREASURY_SECRET_KEY` | Treasury wallet private key | Yes |
| `XDC_RPC_URL` | XDC RPC endpoint | No |
| `JWT_SECRET` | JWT signing secret | Yes |
| `MASTER_ENCRYPTION_KEY` | Encryption key for secrets | Yes |

## Troubleshooting

### Build Fails
- Check build logs in Render dashboard
- Ensure all dependencies are in `go.mod` and `package.json`

### Database Connection Error
- Verify `DATABASE_URL` is correct
- Check if PostgreSQL service is running

### Redis Connection Error
- Verify `REDIS_URL` is correct
- Check if Redis service is running

### API Returns 401
- Ensure `JWT_SECRET` is set
- Check if API key is valid

## Cost Estimate

| Service | Plan | Cost |
|---------|------|------|
| PostgreSQL | Free | $0/month |
| Redis | Free | $0/month |
| Backend API | Free | $0/month |
| Frontend | Free | $0/month |
| **Total** | | **$0/month** |

**Note:** Free tier has limitations:
- Services sleep after 15 minutes of inactivity
- 750 hours/month per service
- 100GB bandwidth/month

## Upgrade to Paid Plan

If you need more resources:
1. Go to each service settings
2. Change plan from **Free** to **Starter** ($7/month per service)
3. Total cost: ~$28/month for all services
