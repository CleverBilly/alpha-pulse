const root = process.env.ALPHA_PULSE_ROOT || "/www/wwwroot/alpha-pulse";
const nodeBin =
  process.env.ALPHA_PULSE_NPM_BIN ||
  "/www/server/nodejs/v24.14.1/bin/npm";

module.exports = {
  apps: [
    {
      name: "alpha-pulse-backend",
      cwd: `${root}/backend`,
      script: `${root}/backend/bin/alpha-pulse`,
      interpreter: "none",
      env: {
        PATH: "/usr/local/btgo/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/root/bin",
      },
      out_file: `${root}/logs/backend.out.log`,
      error_file: `${root}/logs/backend.err.log`,
      autorestart: true,
      max_restarts: 10,
      restart_delay: 3000,
    },
    {
      name: "alpha-pulse-frontend",
      cwd: `${root}/frontend`,
      script: nodeBin,
      args: "run start",
      env: {
        NODE_ENV: "production",
        PATH: "/www/server/nodejs/v24.14.1/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/root/bin",
      },
      out_file: `${root}/logs/frontend.out.log`,
      error_file: `${root}/logs/frontend.err.log`,
      autorestart: true,
      max_restarts: 10,
      restart_delay: 3000,
    },
  ],
};
