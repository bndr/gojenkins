import org.csanchez.jenkins.plugins.kubernetes.*
import org.csanchez.jenkins.plugins.kubernetes.model.*
import jenkins.model.Jenkins
import groovy.json.JsonOutput

{{ if eq .Operation "create" }}
def addKubernetesCloud(cloudList, config) {
    def cloud = new KubernetesCloud(
            cloudName = config.cloudName ?: 'Kubernetes'
    )
    cloud.serverCertificate = config.serverCertificate ?: ''
    cloud.skipTlsVerify = config.skipTlsVerify ?: false
    cloud.credentialsId = config.credentialsId ?: ''
    cloud.jenkinsTunnel = config.jenkinsTunnel ?: ''
    cloud.usageRestricted = config.usageRestricted ?: true
    cloud.serverUrl = config.serverUrl ?: ''
    cloud.namespace = config.namespace ?: ''
    cloud.jenkinsUrl = config.jenkinsUrl ?: ''
    cloud.containerCapStr = config.containerCapStr ?: '10'
    cloudList.add(cloud)
}


private configure(config) {
    def instance = Jenkins.getInstance()
    def clouds = instance.clouds


    config.each { name, details ->
        Iterator iter = clouds.iterator();
        while (iter.hasNext()) {
            elem = iter.next();
            if (elem.name == details.cloudName) {
               iter.remove();
            }
        }
        addKubernetesCloud(clouds, details)
    }

    def lstClouds = []
    clouds.each { elem ->
        def nKubernetesCloudC4 = new KubernetesCloudC4 (
            defaultsProviderTemplate:elem.defaultsProviderTemplate,
            name:                    elem.name,
            serverUrl:               elem.serverUrl,
            useJenkinsProxy:         elem.useJenkinsProxy,
            serverCertificate:       elem.serverCertificate,
            skipTlsVerify:           elem.skipTlsVerify,
            addMasterProxyEnvVars:   elem.addMasterProxyEnvVars,
            capOnlyOnAlivePods:      elem.capOnlyOnAlivePods,
            namespace:               elem.namespace,
            webSocket:               elem.webSocket,
            directConnection:        elem.directConnection,
            jenkinsUrl:              elem.jenkinsUrl,
            jenkinsTunnel:           elem.jenkinsTunnel,
            credentialsId:           elem.credentialsId,
            containerCap:            elem.containerCap,
            retentionTimeout:        elem.retentionTimeout,
            connectTimeout:          elem.connectTimeout,
            readTimeout :            elem.readTimeout ,
            labels:                  elem.labels,
            usageRestricted:         elem.usageRestricted,
            maxRequestsPerHost:      elem.maxRequestsPerHost,
            waitForPodSec :          elem.waitForPodSec,
            podRetention :           elem.podRetention
        )
        def lstLabels = []
        elem.podLabels.each { podLabel ->
            def nLabel = new PodLabel(key: podLabel.key, value: podLabel.value)
            lstLabels.add(nLabel)
        }
        nKubernetesCloudC4.podLabels = lstLabels

        lstClouds.add(nKubernetesCloudC4)
    }
    def lstCloudsJson = JsonOutput.toJson(lstClouds)
    return lstCloudsJson
}
{{ end }}

{{ if eq .Operation "read" }}
private configure(config) {
    def instance = Jenkins.getInstance()
    def clouds = instance.clouds
    def lstClouds = []

    config.each { name, details ->
        Iterator iter = clouds.iterator();
        while (iter.hasNext()) {
            elem = iter.next();
            if (elem.name == details.cloudName) {
                       def nKubernetesCloudC4 = new KubernetesCloudC4 (
                           defaultsProviderTemplate:elem.defaultsProviderTemplate,
                           name:                    elem.name,
                           serverUrl:               elem.serverUrl,
                           useJenkinsProxy:         elem.useJenkinsProxy,
                           serverCertificate:       elem.serverCertificate,
                           skipTlsVerify:           elem.skipTlsVerify,
                           addMasterProxyEnvVars:   elem.addMasterProxyEnvVars,
                           capOnlyOnAlivePods:      elem.capOnlyOnAlivePods,
                           namespace:               elem.namespace,
                           webSocket:               elem.webSocket,
                           directConnection:        elem.directConnection,
                           jenkinsUrl:              elem.jenkinsUrl,
                           jenkinsTunnel:           elem.jenkinsTunnel,
                           credentialsId:           elem.credentialsId,
                           containerCap:            elem.containerCap,
                           retentionTimeout:        elem.retentionTimeout,
                           connectTimeout:          elem.connectTimeout,
                           readTimeout :            elem.readTimeout ,
                           labels:                  elem.labels,
                           usageRestricted:         elem.usageRestricted,
                           maxRequestsPerHost:      elem.maxRequestsPerHost,
                           waitForPodSec :          elem.waitForPodSec,
                           podRetention :           elem.podRetention
                       )
                       def lstLabels = []
                       elem.podLabels.each { podLabel ->
                           def nLabel = new PodLabel(key: podLabel.key, value: podLabel.value)
                           lstLabels.add(nLabel)
                       }
                       nKubernetesCloudC4.podLabels = lstLabels

                       lstClouds.add(nKubernetesCloudC4)
            }
        }
    }


    def lstCloudsJson = JsonOutput.toJson(lstClouds)
    return lstCloudsJson
}
{{ end }}

{{ if eq .Operation "delete" }}
private configure(config) {
    def instance = Jenkins.getInstance()
    def clouds = instance.clouds

    config.each { name, details ->
        Iterator iter = clouds.iterator();
        while (iter.hasNext()) {
            elem = iter.next();
            if (elem.name == details.cloudName) {
               iter.remove();
            }
        }

    }
}
{{ end }}

configure 'k8s-cloud': [
        {{ if or ( eq .Operation "read") (eq .Operation "delete") }}
        cloudName    : '{{ .CloudName }}'
        {{ else if  eq .Operation "create" }}
        cloudName    : '{{ .CloudName }}',
        namespace    : '{{.Namespace}}',
        jenkinsUrl   : '{{ .JenkinsURL}}',
        jenkinsTunnel: '{{ .JenkinsTunnel}}'
        {{ end }}
]


{{ if or ( eq .Operation "create") (eq .Operation "read") }}
public class PodLabel {
    String key;
    String value;
}

public class KubernetesCloudC4 {
    String defaultsProviderTemplate;
    List<PodTemplate> templates;
    String name;
    String serverUrl;
    boolean useJenkinsProxy;
    String serverCertificate;
    boolean skipTlsVerify;
    boolean addMasterProxyEnvVars;
    boolean capOnlyOnAlivePods;
    String namespace;
    boolean webSocket;
    boolean directConnection;
    String jenkinsUrl;
    String jenkinsTunnel;
    String credentialsId;
    Integer containerCap;
    int retentionTimeout;
    int connectTimeout;
    int readTimeout;
    Map<String, String> labels;
    List<PodLabel> podLabels;
    boolean usageRestricted;
    int maxRequestsPerHost;
    Integer waitForPodSec;
    String podRetention;
}
{{ end }}