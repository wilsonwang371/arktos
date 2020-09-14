package handler

import (
	"k8s.io/klog"
	v1 "k8s.io/kubernetes/pkg/apis/cloudgateway/v1"
	clientset "k8s.io/kubernetes/pkg/client/clientset/versioned"
	listers "k8s.io/kubernetes/pkg/client/listers/cloudgateway/v1"
)

// ServiceExposeHandler is a service expose object handler
type ServiceExposeHandler struct {
	serviceLister listers.EServiceLister
	siteLister    listers.ESiteLister
	gatewayClient clientset.Interface

	// Map edge site name to the allocated virtual presence info
	virtualPresenceMap map[string]VirtualPresenceInfo
}

// Allocated VirtualPresence info
type VirtualPresenceInfo struct {
	cidr         string
	allocatedIps []string
}

// ServiceExposeObj
type ServiceExposeObj struct {
	serviceExpose v1.ServiceExpose
	service       v1.EService
	serviceSite   v1.ESite
	clientSite    v1.ESite
}

// NewServiceExposeHandler creates a new ServiceExposeHandler
func NewServiceExposeHandler(serviceLister listers.EServiceLister, siteLister listers.ESiteLister,
	gatewayClient clientset.Interface) *ServiceExposeHandler {
	se := &ServiceExposeHandler{
		serviceLister: serviceLister,
		siteLister:    siteLister,
		gatewayClient: gatewayClient,
		virtualPresenceMap: map[string]VirtualPresenceInfo{},
	}

	// Init map here
	return se
}

// Request a un used virtual presence ip in the network of the site
func (h *ServiceExposeHandler) RequestVirtualPresence(site *v1.ESite, serviceExpose *v1.ServiceExpose) (string, error) {

	return "", nil
}

func (h *ServiceExposeHandler) ReleaseVirtualPresence(site *v1.ESite, serviceExpose *v1.ServiceExpose,
	virtualPresenceIp string) error {
	return nil
}

// Handle the service expose request
// 1. Generate traffic flows from the request
//    traffic flows contains the site and flows basic info
// 2. Send the traffic flows control message to the dataflow model if the associated site is
//    in the cloud, or send it to the associated site from the hub communication tunnel
// 3. The dataflow model can be implemented use the driver/adapter mode, if use transform mechanism data flow
//    by openvswitch, it can work with tap/tun device to do the data flow. In this case, hub communication tunnel
//    must support the Binary Message transfer
func (h *ServiceExposeHandler) ObjectCreated(tenant string, namespace string, obj interface{}) {
	se := obj.(*v1.ServiceExpose)
	klog.V(4).Info("ServiceExposeHandler.ObjectCreated: %v", se)

	// Get the service expose obj
	seObj, err := h.getServiceExposeObj(tenant, namespace, se)
	if err != nil {
		// Update service expose to wrong status and return
		// TODO(nkwangjun): update record to detail message
		nerr := h.updateServiceExposeStatus(se, v1.ServiceExposeError, namespace, tenant)
		if nerr != nil {
			klog.Error("UpdateServiceExpose to error failed, service expose:%v, err:%v", se, nerr)
		}
		return
	}

	// If service virtual presence ip is assigned, do check and update
	if seObj.serviceExpose.VirtualPresenceIp != "" {
		klog.V(4).Infof("ServiceExpose virtual presence ip for service is assigned, se:%v", se)
		// NOTE(nkwangjun): add check and update here later
	} else {
		// Do request virtual presence ip
		klog.V(4).Infof("ServiceExpose try to get one virtual presence ip in site:%v", seObj.clientSite)
		virtualPresenceIp, err := h.RequestVirtualPresence(&seObj.clientSite, se)
		if err != nil {
			klog.Errorf("Request virtual presence ip for service error, se:%v, err:%v", se, err)
			// TODO(nkwangjun): update record to detail message
			nerr := h.updateServiceExposeStatus(se, v1.ServiceExposeError, namespace, tenant)
			if nerr != nil {
				klog.Error("UpdateServiceExpose to error failed, %v, err:%v", se, nerr)
			}
			return
		}

		// Update virtual presence ip for service
		seCopy := seObj.serviceExpose.DeepCopy()
		seCopy.VirtualPresenceIp = virtualPresenceIp
		se, err = h.gatewayClient.CloudgatewayV1().ServiceExposesWithMultiTenancy(namespace, tenant).Update(seCopy)
		if err != nil {
			klog.Errorf("Update virtual presence ip for se error, se:%v, err:%v", se, err)
			// TODO(nkwangjun): update record to detail message
			err = h.ReleaseVirtualPresence(&seObj.clientSite, se, virtualPresenceIp)
			if err != nil {
				klog.Errorf("Release virtual presence for service error, se:%v, err:%v", se, err)
			}
			return
		}
	}

	for _, eClient := range se.AllowedClients {
		// If eClient virtual presence ip is assigned, do check and update
		if eClient.VirtualPresenceIp != "" {
			klog.V(4).Infof("ServiceExpose virtual presence ip for eClient:%s is assigned, se:%v",
				eClient.Ip, se)
			// NOTE(nkwangjun): add check and update here later
		} else {
			// Do request virtual presence ip
			klog.V(4).Infof("ServiceExpose try to get one virtual presence ip in site:%v", seObj.serviceSite)
			virtualPresenceIp, err := h.RequestVirtualPresence(&seObj.serviceSite, se)
			if err != nil {
				klog.Errorf("Request virtual presence ip for service error, se:%v, err:%v", se, err)
				// TODO(nkwangjun): update record to detail message
				nerr := h.updateServiceExposeStatus(se, v1.ServiceExposeError, namespace, tenant)
				if nerr != nil {
					klog.Error("UpdateServiceExpose to error failed, %v, err:%v", se, nerr)
				}
				return
			}

			// Update virtual presence ip for service
			seCopy := se.DeepCopy()
			seCopy.VirtualPresenceIp = virtualPresenceIp
			se, err = h.gatewayClient.CloudgatewayV1().ServiceExposesWithMultiTenancy(namespace, tenant).Update(seCopy)
			if err != nil {
				klog.Errorf("Update virtual presence ip for se error, se:%v, err:%v", se, err)
				// TODO(nkwangjun): update record to detail message
				nerr := h.ReleaseVirtualPresence(&seObj.serviceSite, se, virtualPresenceIp)
				if nerr != nil {
					klog.Errorf("Release virtual presence for service error, se:%v, err:%v", se, nerr)
				}
				return
			}

			klog.V(4).Infof("Request and update virtual presence ip success for eClient, vp:%s, se:%v",
				virtualPresenceIp, se)
		}
	}

	// Update service expose with virtual presence ready
	seObj.serviceExpose = *se

	// Send traffic flow required info
	// 1. Send to the service site
	// 2. Send to the server site
	// 3. Record event every step
	// 4. Update status of the service expose
	// 5. If associated gateway not be assigned, wait until the message send successful
	h.syncServiceExpose(seObj)
}

func (h *ServiceExposeHandler) updateVirtualPresenceForEClient(se *v1.ServiceExpose, eSite *v1.ESite,
	namespace string, tenant string) error {
	for _, eClient := range se.AllowedClients {
		// If eClient virtual presence ip is assigned, do check and update
		if eClient.VirtualPresenceIp != "" {
			klog.V(4).Infof("ServiceExpose virtual presence ip for eClient:%s is assigned, se:%v",
				eClient.Ip, se)
			// NOTE(nkwangjun): add check and update here later
		} else {
			// Do request virtual presence ip
			klog.V(4).Infof("ServiceExpose try to get one virtual presence ip in site:%v", eSite)
			virtualPresenceIp, err := h.RequestVirtualPresence(eSite, se)
			if err != nil {
				klog.Errorf("Request virtual presence ip for service error, se:%v, err:%v", se, err)
				return err
			}

			// Update virtual presence ip for service
			seCopy := se.DeepCopy()
			seCopy.VirtualPresenceIp = virtualPresenceIp
			_, err = h.gatewayClient.CloudgatewayV1().ServiceExposesWithMultiTenancy(namespace, tenant).Update(seCopy)
			if err != nil {
				klog.Errorf("Update virtual presence ip for se error, se:%v, err:%v", se, err)
				// TODO(nkwangjun): update record to detail message
				nerr := h.ReleaseVirtualPresence(eSite, se, virtualPresenceIp)
				if nerr != nil {
					klog.Errorf("Release virtual presence for service error, se:%v, err:%v", se, nerr)
				}
				return err
			}

			klog.V(4).Infof("Request and update virtual presence ip success for eClient, vp:%s, se:%v",
				virtualPresenceIp, se)
		}
	}
	return nil
}

func (h *ServiceExposeHandler) updateServiceExposeStatus(serviceExpose *v1.ServiceExpose,
	phase v1.ServiceExposePhase, namespace string, tenant string) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// We can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	se := serviceExpose.DeepCopy()
	se.Status = v1.ServiceExposeStatus{Phase: phase}
	_, err := h.gatewayClient.CloudgatewayV1().ServiceExposesWithMultiTenancy(namespace, tenant).Update(se)
	return err
}

func (h *ServiceExposeHandler) syncServiceExpose(seObj *ServiceExposeObj) {
	klog.V(4).Infof("Sync service expose:%v", seObj)
}

func (h *ServiceExposeHandler) getServiceExposeObj(tenant string, namespace string, expose *v1.ServiceExpose) (
	*ServiceExposeObj, error) {
	seObj := &ServiceExposeObj{
		serviceExpose: *expose,
	}

	// Check the service by expose
	serviceName := expose.EServiceName
	service, err := h.serviceLister.EServicesWithMultiTenancy(namespace, tenant).Get(serviceName)
	if err != nil {
		klog.Errorf("Get service in service expose error, service expose:%v, service name:%s", expose,
			serviceName)
		return nil, err
	}

	seObj.service = *service

	// Check the site of the service
	serviceSite, err := h.siteLister.ESitesWithMultiTenancy(namespace, tenant).Get(service.ESiteName)
	if err != nil {
		klog.Errorf("Get site from service error, service:%v", service)
		return nil, err
	}

	seObj.serviceSite = *serviceSite

	// Check the site of the client
	clientSite, err := h.siteLister.ESitesWithMultiTenancy(namespace, tenant).Get(expose.ESiteName)
	if err != nil {
		klog.Errorf("Get site from expose client error, expose:%v", expose)
		return nil, err
	}

	seObj.clientSite = *clientSite
	return seObj, nil
}

func (h *ServiceExposeHandler) ObjectDeleted(tenant string, namespace string, obj interface{}) {
	se := obj.(*v1.ServiceExpose)
	klog.V(4).Info("ServiceExposeHandler.ObjectDeleted: %v", se)

	// Transform the service expose to obj
	seObj, err := h.getServiceExposeObj(tenant, namespace, se)
	if err != nil {
		// Update service expose to wrong status and return
		klog.Errorf("Get service expose error:%v", err)
	}

	if seObj == nil {
		return
	}

	// Release policy
	h.releaseServiceExpose(seObj)
}

func (h *ServiceExposeHandler) releaseServiceExpose(seObj *ServiceExposeObj) {
	klog.V(4).Infof("Release service expose:%v", seObj)

	// TODO(nkwangjun): release the policy rules
}
