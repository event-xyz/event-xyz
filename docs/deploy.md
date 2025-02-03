# Deploy Event Loop

0. Clone this repo onto the droplet.
1. For nginx, place `eventloop.hegde.live` at `/etc/nginx/sites-available`.
2. Create symlink: `ln -s /etc/nginx/sites-available/eventloop.hegde.live /etc/nginx/sites-enabled/`
3. Remove symlink for default site: `rm /etc/nginx/sites-enabled/default`
4. Restart nginx: `sudo systemctl restart nginx`
