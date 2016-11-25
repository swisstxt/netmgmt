node('centos7') {
    def name = "netmgmt"
    def branch = env.BRANCH_NAME
    def orgName = "github.com/swisstxt"
    
    def workspaceDir = env.WORKSPACE
    def specsDir = "${workspaceDir}/SPECS"
    def sourcesDir = "${workspaceDir}/SOURCES"
    def relProjectOrgDir = "SOURCES/src/${orgName}"
    def projectOrgDir = "${workspaceDir}/${relProjectOrgDir}"
    def relProjectSourceDir = "${relProjectOrgDir}/${name}"
    def projectSourceDir = "${workspaceDir}/${relProjectSourceDir}"
    def relRpmbuildDir = "rpmbuild"
    def rpmbuildDir = "${workspaceDir}/${relRpmbuildDir}"
    
    def version
    def release = env.BUILD_NUMBER
    def spec
    def arch
    def osRelease
    def rev
    
    stage('Checkout Repo') {
        echo scm.getUserRemoteConfigs().toString()
        checkout scm
    }
    
    stage('Prepare Build') {
        sh "mkdir -p ${projectOrgDir}"
        sh "ln -sf ${workspaceDir} ${projectSourceDir}"
    }
    
    stage('Set Build Variables') {
        spec = sh(
            script: "/opt/buildhelper/buildhelper getspec ${name}",
            returnStdout: true
        ).trim()
        arch = sh(
            script: "/opt/buildhelper/buildhelper getarch || true",
            returnStdout: true
        ).trim()
        osRelease = sh(
            script: "/opt/buildhelper/buildhelper getosrelease || true",
            returnStdout: true
        ).trim()
        version = sh(
            script: "/opt/buildhelper/buildhelper getgittag ${workspaceDir}",
            returnStdout: true
        ).trim()
        rev = sh(
            script: "git rev-parse --short HEAD",
            returnStdout: true
        ).trim()
        env.GOPATH = sourcesDir
        env.PATH = "${sourcesDir}/bin:${env.PATH}"
    }
    
    stage('Clean Before Build') {
        for (path in [rpmbuildDir, "${sourcesDir}/pkg", "${sourcesDir}/lib"]) {
            sh "rm -rf ${path}"
        }
        sh "rm -f ${sourcesDir}/netmgmt.bin"
        for (path in ["${rpmbuildDir}/SPECS", "${rpmbuildDir}/SOURCES", specsDir, "${sourcesDir}/pkg", "${sourcesDir}/lib"]) {
            sh "mkdir -p ${path}"
        }
    }
    
    stage('Get Dependencies') {
        sh "go get github.com/tools/godep"
    }
    
    stage('Compile Source') {
        dir(projectSourceDir) {
            sh """
                cd ${projectSourceDir}
                pwd
                godep restore
                godep go install
            """
        }
        sh "cp ${sourcesDir}/bin/netmgmt ${sourcesDir}/netmgmt.bin"
    }
    
    stage('Build RPM') {
        sh "cp -r ${specsDir}/* ${rpmbuildDir}/SPECS/ || true"
        sh "cp -r ${sourcesDir}/* ${rpmbuildDir}/SOURCES/ || true"
        sh """
            rpmbuild -ba ${spec} \
            --define "ver ${version}" \
            --define "rel ${release}" \
            --define "name ${name}" \
            --define "os_rel ${osRelease}" \
            --define "arch ${arch}" \
            --define "_topdir ${rpmbuildDir}" \
            --define "_builddir %{_topdir}" \
            --define "_rpmdir %{_topdir}" \
            --define "_srcrpmdir %{_topdir}"
        """
    }
    
    stage('Archive RPM') {
        archiveArtifacts "${relRpmbuildDir}/*.rpm"
        archiveArtifacts "${relRpmbuildDir}/*/*.rpm"
    }
}
