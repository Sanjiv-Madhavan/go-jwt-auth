package controllers

import "net/http"

func (c *Controller) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StateActive
	healthStatus := []string{"Hello", "Active"}
	c.logger.Info(http.StatusText(int(status)))
	c.middleware.SendJSONResponse(w, int(status), healthStatus)
	return
}
