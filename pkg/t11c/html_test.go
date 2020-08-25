package t11c

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractWANIP(t *testing.T) {
	// Trimmed and anonymised copy of statusview.cgi response content
	const testCorrectIP = "192.0.2.138"
	const statusViewBody = `
<html><head><meta http-equiv="Content-Type" content="text/html hhh; charset=iso-8859-1"></head><body>
<div class="title" style="color:#CCC;"><span id="SystemInfo_ConnectionStatus"></span></div>
<table width="96%" cellspacing="0" cellpadding="0" border="0" align="left">
<tbody>
<tr>
<td valign="top"><div class="w_text3">
<table class="table_frame" width="96%" cellspacing="0" cellpadding="0" border="0" align="center">
<tbody>
    <tr>
    <td class="table_font">&nbsp;&nbsp;-  <span id="MLG_IP_Address2"></span>: </td>
    <td class="table_font w_blue" id="DeviceInfo_WanIP">
192.0.2.138&nbsp;&nbsp;<input type="button" name="Disconnect" maxlength="32" value="Disconnect" onclick="reconnect(2)">
</td>
    </tr>
    <tr>
    <td class="table_font">&nbsp;&nbsp;- <span id="MLG_IP_Subnet_Mask"></span>:</td>
    <td class="table_font w_blue" id="DeviceInfo_WanSubMask">
255.255.255.255
</td>
    </tr>
    <tr>
    <td class="table_font">&nbsp;&nbsp;- <span id="MLG_Default_Gateway"></span>:</td>
    <td class="table_font w_blue" id="DeviceInfo_gateway">
198.51.100.200
</td>
    </tr>
</tbody></table>
</div></td>
</tr></tbody></table>
</body></html>`

	ip, err := extractWANIP(strings.NewReader(statusViewBody))
	assert.NoError(t, err, "Should retrieve the IP from the connected body without error")
	assert.Equal(t, testCorrectIP, ip, "Should extract the correct, trimmed, IP")

	// Dummy response content similar to an invalid session
	const otherBody = `
<html><head<
<title></title>
<meta http-equiv="Cache-Control" CONTENT="no-cache">
</head>
<body></body>
<script language="JavaScript">
top.location.href = "http://192.168.1.1/cgi-bin/login.html";
</script>
</html>`

	_, err = extractWANIP(strings.NewReader(otherBody))
	assert.Error(t, err, "Should error if the WAN IP element does not exist")
	assert.Equal(t, errWANIPElementNotFound, err, "Error from element not exists should match sentinel value")

	// Status view content without IP text
	missingTextBody := strings.Replace(statusViewBody, testCorrectIP, "", -1)

	_, err = extractWANIP(strings.NewReader(missingTextBody))
	assert.Error(t, err, "Should error if the WAN IP element does not contain an IP")
	assert.Equal(t, errWANIPTextNotFound, err, "Error from IP not in element should match sentinel value")
}
