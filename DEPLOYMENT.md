# Render.com Deployment Guide

## ✅ Pre-deployment Checklist - COMPLETE!

- [x] `render.yaml` configuration file created
- [x] `.dockerignore` file created for optimized builds
- [x] `README.md` updated with deployment instructions
- [x] Docker configuration tested locally
- [x] All changes committed to Git
- [x] Changes pushed to GitHub (develop branch)

## 🚀 Next Steps: Deploy to Render.com

### Option 1: Deploy from Master (Recommended for Production)

Follow Git Flow release process:

```bash
# 1. Create release branch from develop
git checkout develop
git pull origin develop
git checkout -b release/1.0.0

# 2. Optional: Make any last-minute release tweaks
# (version numbers, changelog updates, etc.)
git commit -m "Prepare v1.0.0 release"

# 3. Merge to master
git checkout master
git pull origin master
git merge --no-ff release/1.0.0
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin master
git push origin v1.0.0

# 4. Merge back to develop
git checkout develop
git merge --no-ff release/1.0.0
git push origin develop

# 5. Delete release branch
git branch -d release/1.0.0
```

Then deploy from **master** branch on Render.com.

### Option 2: Deploy from Develop (For Testing)

You can deploy directly from develop branch for testing purposes.

## 📋 Render.com Deployment Steps

1. **Sign Up / Log In**
   - Go to https://render.com
   - Sign up with GitHub account
   - Verify email

2. **Create New Blueprint**
   - Click "New +" → "Blueprint"
   - Select repository: `ldvorski/treblle_shiphappens`
   - Choose branch: `master` (after release) or `develop` (for testing)
   - Render will detect `render.yaml` automatically
   - Click "Apply"

3. **Wait for Build**
   - Initial build takes 3-5 minutes
   - Watch the logs for any errors
   - Build steps:
     - Pulling Docker base images
     - Installing dependencies
     - Compiling Go application
     - Creating final image
     - Starting container

4. **Verify Deployment**
   - Once deployed, you'll get a URL: `https://trebble-api-monitor.onrender.com`
   - Test endpoints:
     ```bash
     # Health check
     curl https://trebble-api-monitor.onrender.com/health
     
     # Swagger docs
     https://trebble-api-monitor.onrender.com/swagger/index.html
     
     # Test Jikan proxy
     curl https://trebble-api-monitor.onrender.com/api/jikan/anime/1
     
     # View requests
     curl https://trebble-api-monitor.onrender.com/api/requests
     ```

## 🔧 Configuration Details

### Environment Variables (Already in render.yaml)
- `GIN_MODE=release` - Production mode
- `DB_PATH=/var/data/api_monitor.db` - Database location
- `TZ=UTC` - Timezone
- `PORT=8080` - Application port

### Persistent Disk
- **Name**: database
- **Mount Path**: `/var/data`
- **Size**: 1 GB
- **Purpose**: SQLite database persistence

### Health Check
- **Path**: `/health`
- **Expected**: 200 OK response
- **Purpose**: Container health monitoring

## ⚠️ Important Notes

### Free Tier Limitations
1. **Cold Starts**: App spins down after 15 min of inactivity
   - First request after sleep: 15-30 second delay
   - Solution: Use uptime monitoring service (see below)

2. **Build Time**: Slower on free tier
   - Initial: 3-5 minutes
   - Subsequent: 1-2 minutes (with cache)

3. **Resources**: 
   - 512 MB RAM
   - Shared CPU
   - 100 GB bandwidth/month

### Keep Your App Awake (Optional)

Use a free uptime monitoring service to ping your app every 10 minutes:

**Option 1: UptimeRobot** (https://uptimerobot.com)
- Free tier: 50 monitors
- Interval: 5 minutes
- Ping URL: `https://trebble-api-monitor.onrender.com/health`

**Option 2: cron-job.org** (https://cron-job.org)
- Free tier: Unlimited jobs
- Interval: 1 minute minimum
- URL: `https://trebble-api-monitor.onrender.com/health`

**Option 3: GitHub Actions** (free with your repo)
Create `.github/workflows/keepalive.yml`:
```yaml
name: Keep Alive
on:
  schedule:
    - cron: '*/10 * * * *'  # Every 10 minutes
jobs:
  ping:
    runs-on: ubuntu-latest
    steps:
      - name: Ping health endpoint
        run: curl https://trebble-api-monitor.onrender.com/health
```

## 🎯 Post-Deployment Checklist

- [ ] Service deployed successfully on Render
- [ ] Health check passes
- [ ] API endpoints responding correctly
- [ ] Database is persisting data across requests
- [ ] Swagger documentation accessible
- [ ] Set up uptime monitoring (optional)
- [ ] Add custom domain (optional)
- [ ] Monitor logs in Render dashboard

## 📊 Monitoring Your Deployment

### View Logs
- Go to Render dashboard
- Select your service
- Click "Logs" tab
- Watch real-time logs

### Check Metrics
- Dashboard shows:
  - CPU usage
  - Memory usage
  - Request count
  - Bandwidth usage

### Troubleshooting

**Build fails:**
- Check Dockerfile syntax
- Verify go.mod dependencies
- Review build logs in Render

**App crashes on startup:**
- Check logs for errors
- Verify environment variables
- Ensure port 8080 is exposed

**Database not persisting:**
- Verify disk is mounted at `/var/data`
- Check DB_PATH environment variable
- Review disk settings in render.yaml

**Slow first request:**
- Normal on free tier (cold start)
- Set up uptime monitoring to prevent sleep

## 🔗 Useful Links

- **Render Dashboard**: https://dashboard.render.com
- **Render Docs**: https://render.com/docs
- **Your Repository**: https://github.com/ldvorski/treblle_shiphappens
- **Render Status**: https://status.render.com

## 💡 Next Steps After Deployment

1. **Test all endpoints** thoroughly
2. **Set up uptime monitoring** if needed
3. **Add custom domain** (optional)
4. **Monitor usage** to stay within free tier
5. **Consider upgrading** if you exceed limits

## 🎉 You're Ready to Deploy!

Your application is fully prepared for Render.com deployment. Follow the steps above and you'll be live in minutes!

Good luck! 🚀

