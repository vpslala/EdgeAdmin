Tea.context(function () {
	this.updateItem = function (itemId) {
		teaweb.popup(Tea.url(".updateIPPopup?firewallPolicyId=" + this.firewallPolicyId, {itemId: itemId}), {
			height: "26em",
			callback: function () {
				teaweb.success("保存成功", function () {
					teaweb.reload()
				})
			}
		})
	}

	this.deleteItem = function (itemId) {
		let that = this
		teaweb.confirm("确定要删除这个IP吗？", function () {
			that.$post(".deleteIP")
				.params({
					"firewallPolicyId": this.firewallPolicyId,
					"itemId": itemId
				})
				.refresh()
		})
	}

	/**
	 * 添加IP名单菜单
	 */
	this.createIP = function (type) {
		let that = this
		teaweb.popup("/servers/components/waf/ipadmin/createIPPopup?firewallPolicyId=" + this.firewallPolicyId + '&type=' + type, {
			height: "26em",
			callback: function () {
				teaweb.success("保存成功", function () {
					window.location = "/servers/components/waf/ipadmin/lists?firewallPolicyId=" + that.firewallPolicyId + "&type=" + type
				})
			}
		})
	}
})