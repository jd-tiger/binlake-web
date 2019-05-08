$.get("/admin/getCluster", function (data) {
    data = data["rows"]
    var html = "<option>" + "请选择" + "</option>";
    for (var i = 0; i < data.length; i++) {
        html += "<option value='" + data[i]["id"] + "'>" + data[i]["name"] + "</option>"
    }
    $("#cluster").empty();
    $("#cluster").append(html);
});

$("#cluster").change(function () {
    var cluster = $("#cluster").val();
    $.get("/admin/getWave?cluster=" + cluster, function (data) {
    data = data["rows"]
        var html = "<option>" + "请选择" + "</option>";
        html += "<option value='" + "all" + "'>" + "全选" + "</option>"
        for (var i = 0; i < data.length; i++) {
            html += "<option value='" + data[i]["ip"] + "'>" + data[i]["ip"] + "</option>"
        }
        $("#wave").empty();
        $("#wave").append(html);
    });
});

$(document).ready(function () {
    $('#mysql-table-content').hide();
    $('#wave-table-content').hide();
});

function getIdSelections() {
    return $.map($("#table-mysql").bootstrapTable('getSelections'), function (row) {
        return row.host + ':' + row.port
    });
}

$("#queryMysqlBtn").click(function () {
    var cluster = $("#cluster").val();
    var wave = $("#wave").val();

    if(cluster=='请选择') {
        showMessage("请选择集群");
    } else {
        if(wave=='请选择') {
            initWaveTable();
        } else {
            initMysqlTable();
       }
    }
});

function initWaveTable() {
    $('#table-wave').bootstrapTable('destroy');
    $('#mysql-table-content').hide();
    $('#wave-table-content').show();

    $("#table-wave").bootstrapTable({
        columns:[
        {
            field:'ip',
            title:'ip',
            sortable:true
        },
        {
            field:'cluster_name',
            title:'cluster_name',
            sortable:true
        },
        {
            field:'mysql_num',
            title:'mysql_num',
            sortable:true
        },
        {
            field:'update_time',
            title:'update_time',
            sortable:true
        },
        {
            field:'create_time',
            title:'create_time',
            sortable:true
        }],
        method: "get",  //使用get请求到服务器获取数据
        url: "/admin/getWaveDetail", //获取数据的Servlet地址
        striped: true,  //表格显示条纹
        pagination: false, //启动分页
        clickToSelect:true,
        pageSize: 16,  //每页显示的记录数
        pageNumber:1, //当前第几页
        pageList: [15,20,30,45],  //记录数可选列表
        // search: true,  //是否启用查询
        // showColumns: true,  //显示下拉框勾选要显示的列
        // showRefresh: true,  //显示刷新按钮
        sidePagination: "server", //表示服务端请求
        //设置为undefined可以获取pageNumber，pageSize，searchText，sortName，sortOrder
        //设置为limit可以获取limit, offset, search, sort, order
        queryParamsType : "limit",
        queryParams: function queryParams(params) {   //设置查询参数
            var param = {
                /*limit: params.limit,
                offset: params.offset,
                sort:params.sort,
                order:params.order,*/
                /*host:$("#query-host").val(),
                appName:$("#query-appname").val()*/
                cluster:$("#cluster").val(),
            };
            return param;
        },
        // onLoadSuccess: function(){  //加载成功时执行
        //     showMessage("加载成功");
        // },
        onLoadError: function(){  //加载失败时执行
            showMessage("加载数据失败", {time : 1500, icon : 2});
        }
    });
}

function initMysqlTable() {
    $('#table-mysql').bootstrapTable('destroy');
    $('#wave-table-content').hide();
    $('#mysql-table-content').show();

    $("#table-mysql").bootstrapTable({
        columns:[
        {
            field: 'state',//必须是state，否则无法取到id
            checkbox: true,
            align: 'center',
            valign: 'middle'
        },
        {
            field:'host',
            title:'host',
            sortable:true
        },
        {
            field:'port',
            title:'port',
            sortable:true
            },
            {
            field:'nodeState',
            title:'nodeState',
            sortable:true,
            formatter:function (v,r,i) {
                //这里有5个取值代表5中颜色['active', 'success', 'info', 'warning', 'danger'];
                var classesValue = "";
                if(v == 'ONLINE'){
                    v = '<span class="label label-info">开启</span>';
                }else if(v == 'OFFLINE'){
                    v = '<span class="label label-danger">关闭</span>';
                }
                return v;
            }
        },
        {
            field:'binlogFile',
            title:'binlogFile',
            sortable:true
        },
        {
            field:'binlogPos',
            title:'binlogPos',
            sortable:true
        },
        {
            field:'executedGtidSets',
            title:'executedGtidSets',
            sortable:true
        },
        {
            field:'leader',
            title:'leader',
            sortable:true
        },
        {
            field:'retryTimes',
            title:'retryTimes',
            sortable:true
        },
        {
            field: 'operate',
            title: 'operate',
            align: 'center',
            //events: operateEvents,
            formatter: operateFormatter
        }],
        method: "get",  //使用get请求到服务器获取数据
        url: "/admin/getMysqlDetail", //获取数据的Servlet地址
        striped: true,  //表格显示条纹
        pagination: false, //启动分页
        clickToSelect:true,
        pageSize: 16,  //每页显示的记录数
        pageNumber:1, //当前第几页
        pageList: [15,20,30,45],  //记录数可选列表
        // search: true,  //是否启用查询
        // showColumns: true,  //显示下拉框勾选要显示的列
        // showRefresh: true,  //显示刷新按钮
        sidePagination: "server", //表示服务端请求
        //设置为undefined可以获取pageNumber，pageSize，searchText，sortName，sortOrder
        //设置为limit可以获取limit, offset, search, sort, order
        queryParamsType : "limit",
        queryParams: function queryParams(params) {   //设置查询参数
            var param = {
                cluster:$("#cluster").val(),
                wave:$("#wave").val()
            };
            return param;
        },
        // onLoadSuccess: function(){  //加载成功时执行
        //     showMessage("加载成功");
        // }
        onLoadError: function(){  //加载失败时执行
            showMessage("加载数据失败", {time : 1500, icon : 2});
        }
    });
}

function operateFormatter(value, row, index) {
    return [
        '<button id="transferLeaderBtn" type="button" class="btn btn-danger" style="margin-right:15px;" onclick="showDialog()">迁移Leader</button>',
        '<button id="showSnapshot" type="button" class="btn btn-danger" style="margin-right:15px;" onclick="showSnapshotDialog()">数据回滚</button>'
    ].join('');
}

$("#setLeaderIp").on('click',function () {
    setLeader()
});

function setLeader() {
    var mysqls = getIdSelections();
    var clusterId = $("#cluster").val();
    var leaderIps = $("#leader-ip").val();
    if(leaderIps=='') {
        showMessage("请输入IP");
    } else {
        var params = {
            "mysqls":mysqls,
            "clusterId":clusterId,
            "leaderIps":leaderIps
        };
        $.ajax({
            type:"post",
            dataType:"json",
            url:"/admin/setLeader",
            data:JSON.stringify(params),
            success:function (data) {
                if(data != undefined && data.code == 1000){
                    showMessage("成功！");
                    initMysqlTable();
                }else {
                    showMessage(data.message)
                }
            }
        });
    }
}

$("#btn-reset-counter").on('click',function () {
   doReset("counter")
});

$("#btn-reset-position").on('click',function () {
   doReset("position")
});

function doReset(op) {
    var mysqls = getIdSelections();
    var clusterId = $("#cluster").val();
    console.log(mysqls);
    if(mysqls == ''){
        showMessage("请选择数据");
    }else {
        confirmDialog(mysqls,"确认","确认reset ?" + op,function () {
            var params = {
                "mysqls":mysqls,
                "clusterId":clusterId,
                "op":op
            };
            $.ajax({
                type:"post",
                dataType:"json",
                url:"/admin/reset",
                data:JSON.stringify(params),
                success:function (data) {
                    if(data != undefined && data.code == 1000){
                        showMessage("成功！");
                        initMysqlTable();
                    }else {
                        showMessage(data.message)
                    }
                }
            });
        })
    }
}


function doSwitch(op) {
    var mysqls = getIdSelections();
    var clusterId = $("#cluster").val();
    if(mysqls == ''){
        showMessage("请选择数据");
    }else{
        confirmDialog(mysqls,"确认","设置为" + op,function () {
            var params = {
                "mysqls":mysqls,
                "clusterId":clusterId,
                "op":op
            };
            $.ajax({
                type:"post",
                dataType:"json",
                url:"/admin/switch",
                data:JSON.stringify(params),
                success:function (data) {
                    if(data != undefined && data.code == 1000){
                        showMessage("成功！");
                        initMysqlTable();
                    }else {
                        showMessage("<h3>" + data.message + "</h3>")
                    }
                }
            });
        })
    }
}

$("#btn-online").on('click',function () {
   doSwitch("online")
});

$("#btn-offline").on('click',function () {
    doSwitch("offline")
});

$("#querySnapshotBtn").on('click',function () {
    getSnapshot()
});

$("#btn-selectSnapshot").on('click',function () {
    setSnapshot()
});

function showDialog(content) {
    if (content == "<h3></h3>Forbidden") {
        top.location.reload();
        $("#message_modal_content").html("<h3>会话超时，重新登录</h3>");
        $("#message_dialog").modal('show');
        return;
    }
    //$("#mysql_modal_content").html(content);
    $("#mysql_dialog").modal('show');
}

function showSnapshotDialog(content) {
    if (content == "<h3></h3>Forbidden") {
        top.location.reload();
        $("#message_modal_content").html("<h3>会话超时，重新登录</h3>");
        $("#message_dialog").modal('show');
        return;
    }
    $("#snapshot_dialog").modal('show');

}

function getSnapshot() {
    var mysqls = getIdSelections();
    var params = {
        "mysqls":mysqls
    };
    $.ajax({
        type:"post",
        dataType:"json",
        url:"/admin/getSnapshot",
        data:JSON.stringify(params),
        success:function (data) {
            if(data.code == 150){
                showMessage("<h3>" + data.message + "</h3>")
            }
            data = data["rows"]
            var html = "<option>" + "请选择" + "</option>";
            for (var i = 0; i < data.length; i++) {
                html += "<option value='" + data[i]["create_time"] + "'>" + data[i]["create_time"] + "</option>"
            }
            $("#snapshot_time").empty();
            $("#snapshot_time").append(html);
        }
    });
}

function setSnapshot() {
    var mysqls = getIdSelections();
    var time = $("#snapshot_time").val();
    var clusterId = $("#cluster").val();
    if(time == '请选择'){
        showMessage("请选择时间");
    }else{
        var params = {
            "clusterId":clusterId,
            "mysqls":mysqls,
            "time":time
        };
        $.ajax({
            type:"post",
            dataType:"json",
            url:"/admin/setSnapshot",
            data:JSON.stringify(params),
            success:function (data) {
                if(data != undefined && data.code == 1000){
                    showMessage("成功！");
                    initMysqlTable();
                }else {
                    showMessage("<h3>" + data.message + "</h3>")
                }
            }
        });
    }
}
