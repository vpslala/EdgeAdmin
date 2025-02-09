Tea.context(function () {
    let isInstalling = false

    this.$delay(function () {
        this.reloadStatus(this.nodeId)
    })

    // 开始安装
    this.install = function () {
        isInstalling = true

        this.$post("$")
            .params({
                nodeId: this.nodeId
            })
            .success(function () {

            })
    }

    // 设置节点安装状态
    this.updateNodeIsInstalled = function (isInstalled) {
        teaweb.confirm("确定要将当前节点修改为未安装状态？", function () {
            this.$post("/ns/clusters/cluster/node/updateInstallStatus")
                .params({
                    nodeId: this.nodeId,
                    isInstalled: isInstalled ? 1 : 0
                })
                .refresh()
        })
    }

    // 刷新状态
    this.reloadStatus = function (nodeId) {
        let that = this

        this.$post("/ns/clusters/cluster/node/status")
            .params({
                nodeId: nodeId
            })
            .success(function (resp) {
                this.installStatus = resp.data.installStatus
                this.node.isInstalled = resp.data.isInstalled

                if (!isInstalling) {
                    return
                }

                let nodeId = this.node.id
                let errMsg = this.installStatus.error

                if (this.installStatus.errorCode.length > 0) {
                    isInstalling = false
                }

                switch (this.installStatus.errorCode) {
                    case "EMPTY_LOGIN":
                    case "EMPTY_SSH_HOST":
                    case "EMPTY_SSH_PORT":
                    case "EMPTY_GRANT":
                        teaweb.warn("需要填写SSH登录信息", function () {
                            teaweb.popup("/ns/clusters/cluster/updateNodeSSH?nodeId=" + nodeId, {
                                callback: function () {
                                    that.install()
                                }
                            })
                        })
                        return
                    case "SSH_LOGIN_FAILED":
                        teaweb.warn("SSH登录失败，请检查设置", function () {
                            teaweb.popup("/ns/clusters/cluster/updateNodeSSH?nodeId=" + nodeId, {
                                callback: function () {
                                    that.install()
                                }
                            })
                        })
                        return
                    case "CREATE_ROOT_DIRECTORY_FAILED":
                        teaweb.warn("创建根目录失败，请检查目录权限或者手工创建：" + errMsg)
                        return
                    case "INSTALL_HELPER_FAILED":
                        teaweb.warn("安装助手失败：" + errMsg)
                        return
                    case "TEST_FAILED":
                        teaweb.warn("环境测试失败：" + errMsg)
                        return
                    case "RPC_TEST_FAILED":
                        teaweb.confirm("html:要安装的节点到API服务之间的RPC通讯测试失败，具体错误：" + errMsg + "，<br/>现在修改API信息？", function () {
                            window.location = "/api"
                        })
                        return
                    default:
                        shouldReload = true
                    //teaweb.warn("安装失败：" + errMsg)
                }
            })
            .done(function () {
                this.$delay(function () {
                    this.reloadStatus(nodeId)
                }, 1000)
            });
    }
})