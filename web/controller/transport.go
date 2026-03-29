package controller

import (
	"github.com/mhsanaei/3x-ui/v2/web/service"

	"github.com/gin-gonic/gin"
)

type TransportController struct {
	transportService service.ManagedTransportService
}

type transportConfigPayload struct {
	Content string `json:"content" form:"content"`
}

type trustTunnelConfigPayload struct {
	ListenHost      string `json:"listenHost" form:"listenHost"`
	ListenPort      int    `json:"listenPort" form:"listenPort"`
	CredentialsFile string `json:"credentialsFile" form:"credentialsFile"`
	Hostname        string `json:"hostname" form:"hostname"`
	CertChainPath   string `json:"certChainPath" form:"certChainPath"`
	PrivateKeyPath  string `json:"privateKeyPath" form:"privateKeyPath"`
	PublicAddress   string `json:"publicAddress" form:"publicAddress"`
}

type mtprotoConfigPayload struct {
	BindHost   string `json:"bindHost" form:"bindHost"`
	Port       int    `json:"port" form:"port"`
	Secret     string `json:"secret" form:"secret"`
	ConfigPath string `json:"configPath" form:"configPath"`
}

type trustTunnelClientPayload struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

func NewTransportController(g *gin.RouterGroup) *TransportController {
	a := &TransportController{}
	a.initRouter(g)
	return a
}

func (a *TransportController) initRouter(g *gin.RouterGroup) {
	g.GET("/services", a.servicesPage)
	api := g.Group("/api/services")
	api.GET("/status", a.getStatuses)
	api.GET("/trusttunnel", a.getTrustTunnelConfig)
	api.GET("/trusttunnel/cert-paths", a.detectTrustTunnelCertificatePaths)
	api.POST("/trusttunnel", a.saveTrustTunnelConfig)
	api.GET("/trusttunnel/clients", a.getTrustTunnelClients)
	api.POST("/trusttunnel/clients", a.addTrustTunnelClient)
	api.POST("/trusttunnel/clients/:username/delete", a.deleteTrustTunnelClient)
	api.GET("/trusttunnel/clients/:username/export", a.exportTrustTunnelClient)
	api.GET("/mtproto", a.getMTProtoConfig)
	api.POST("/mtproto", a.saveMTProtoConfig)
	api.GET("/configs", a.getConfigs)
	api.GET("/config/:key", a.getConfig)
	api.POST("/config/:key", a.saveConfig)
	api.POST("/:key/:action", a.runAction)
}

func (a *TransportController) servicesPage(c *gin.Context) {
	html(c, "services.html", "Services", nil)
}

func (a *TransportController) getStatuses(c *gin.Context) {
	jsonObj(c, a.transportService.ListStatuses(), nil)
}

func (a *TransportController) getTrustTunnelConfig(c *gin.Context) {
	item, err := a.transportService.GetTrustTunnelConfig()
	if err != nil {
		jsonMsg(c, "Failed to read TrustTunnel config", err)
		return
	}
	jsonObj(c, item, nil)
}

func (a *TransportController) detectTrustTunnelCertificatePaths(c *gin.Context) {
	item, err := a.transportService.DetectTrustTunnelCertificatePaths(c.Query("hostname"))
	if err != nil {
		jsonMsg(c, "Failed to detect TrustTunnel certificate paths", err)
		return
	}
	jsonObj(c, item, nil)
}

func (a *TransportController) saveTrustTunnelConfig(c *gin.Context) {
	payload := &trustTunnelConfigPayload{}
	if err := c.ShouldBind(payload); err != nil {
		jsonMsg(c, "Failed to parse TrustTunnel config", err)
		return
	}
	if err := a.transportService.SaveTrustTunnelConfig(&service.TrustTunnelServiceConfig{
		ListenHost:      payload.ListenHost,
		ListenPort:      payload.ListenPort,
		CredentialsFile: payload.CredentialsFile,
		Hostname:        payload.Hostname,
		CertChainPath:   payload.CertChainPath,
		PrivateKeyPath:  payload.PrivateKeyPath,
		PublicAddress:   payload.PublicAddress,
	}); err != nil {
		jsonMsg(c, "Failed to save TrustTunnel config", err)
		return
	}
	item, err := a.transportService.GetTrustTunnelConfig()
	if err != nil {
		jsonMsg(c, "TrustTunnel config saved but reload failed", err)
		return
	}
	jsonMsgObj(c, "TrustTunnel config saved", item, nil)
}

func (a *TransportController) getTrustTunnelClients(c *gin.Context) {
	items, err := a.transportService.ListTrustTunnelClients()
	if err != nil {
		jsonMsg(c, "Failed to read TrustTunnel clients", err)
		return
	}
	jsonObj(c, items, nil)
}

func (a *TransportController) addTrustTunnelClient(c *gin.Context) {
	payload := &trustTunnelClientPayload{}
	if err := c.ShouldBind(payload); err != nil {
		jsonMsg(c, "Failed to parse TrustTunnel client payload", err)
		return
	}
	item, err := a.transportService.AddTrustTunnelClient(payload.Username, payload.Password)
	if err != nil {
		jsonMsg(c, "Failed to create TrustTunnel client", err)
		return
	}
	jsonMsgObj(c, "TrustTunnel client created", item, nil)
}

func (a *TransportController) deleteTrustTunnelClient(c *gin.Context) {
	if err := a.transportService.DeleteTrustTunnelClient(c.Param("username")); err != nil {
		jsonMsg(c, "Failed to delete TrustTunnel client", err)
		return
	}
	jsonMsg(c, "TrustTunnel client deleted", nil)
}

func (a *TransportController) exportTrustTunnelClient(c *gin.Context) {
	link, err := a.transportService.ExportTrustTunnelClient(c.Param("username"))
	if err != nil {
		jsonMsg(c, "Failed to export TrustTunnel client", err)
		return
	}
	jsonObj(c, gin.H{
		"deeplink": link,
	}, nil)
}

func (a *TransportController) getMTProtoConfig(c *gin.Context) {
	item, err := a.transportService.GetMTProtoConfig()
	if err != nil {
		jsonMsg(c, "Failed to read MTProto config", err)
		return
	}
	jsonObj(c, item, nil)
}

func (a *TransportController) saveMTProtoConfig(c *gin.Context) {
	payload := &mtprotoConfigPayload{}
	if err := c.ShouldBind(payload); err != nil {
		jsonMsg(c, "Failed to parse MTProto config", err)
		return
	}
	if err := a.transportService.SaveMTProtoConfig(&service.MTProtoServiceConfig{
		BindHost:   payload.BindHost,
		Port:       payload.Port,
		Secret:     payload.Secret,
		ConfigPath: payload.ConfigPath,
	}); err != nil {
		jsonMsg(c, "Failed to save MTProto config", err)
		return
	}
	item, err := a.transportService.GetMTProtoConfig()
	if err != nil {
		jsonMsg(c, "MTProto config saved but reload failed", err)
		return
	}
	jsonMsgObj(c, "MTProto config saved", item, nil)
}

func (a *TransportController) getConfigs(c *gin.Context) {
	jsonObj(c, a.transportService.ListConfigs(), nil)
}

func (a *TransportController) getConfig(c *gin.Context) {
	item, err := a.transportService.GetConfig(c.Param("key"))
	if err != nil {
		jsonMsg(c, "Failed to read config", err)
		return
	}
	jsonObj(c, item, nil)
}

func (a *TransportController) saveConfig(c *gin.Context) {
	payload := &transportConfigPayload{}
	if err := c.ShouldBind(payload); err != nil {
		jsonMsg(c, "Failed to parse config payload", err)
		return
	}
	if err := a.transportService.SaveConfig(c.Param("key"), payload.Content); err != nil {
		jsonMsg(c, "Failed to save config", err)
		return
	}
	item, err := a.transportService.GetConfig(c.Param("key"))
	if err != nil {
		jsonMsg(c, "Config saved but reload failed", err)
		return
	}
	jsonMsgObj(c, "Config saved", item, nil)
}

func (a *TransportController) runAction(c *gin.Context) {
	key := c.Param("key")
	action := c.Param("action")
	err := a.transportService.RunAction(key, action)
	if err != nil {
		jsonMsg(c, "Service action failed", err)
		return
	}
	jsonMsgObj(c, "Service action executed", a.transportService.ListStatuses(), nil)
}
