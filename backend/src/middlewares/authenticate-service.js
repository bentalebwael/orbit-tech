const { env } = require("../config");
const { authenticateToken } = require("./authenticate-token");

const authenticateService = (req, res, next) => {
  const apiKey = req.headers['x-api-key'];

  // Check if request is from internal service
  if (apiKey && apiKey === env.INTERNAL_SERVICE_API_KEY) {
    // Set service context and bypass user authentication
    req.user = {
      id: 'service',
      role: 'service',
      type: 'internal'
    };
    return next();
  }

  // Fall back to regular user authentication for other requests
  return authenticateToken(req, res, next);
};

module.exports = { authenticateService };
