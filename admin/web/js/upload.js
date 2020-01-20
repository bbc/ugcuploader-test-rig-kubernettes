
$(document).ready(function () {

    var selected = []
    var tenant
    $('a.dropdown-item').on('click', function (e) {
        e.preventDefault();
        tenant = $(this).text();

        table = $('#tenantReport').DataTable({
            "dom": 'Blrtip',
            processing: false,
            serverSide: false,
            select: true,
            destroy: true,
            columnDefs: [{
                "targets": 0,
                "className": 'select-checkbox'
            }],
            columns: [
                {
                    data: null,
                    defaultContent: '',
                    className: 'select-checkbox',
                    orderable: false
                },
                { data: 'date' },
            ],
            select: {
                style: 'multi',
                selector: 'td:first-child'
            },
            order: [[0, 'asc']],
            buttons: [
                {
                    text: 'Create Report for ' + tenant,
                    action: function (e, dt, node, config) {

                        var data = table.rows({ selected: true }).data();
                        var newarray = [];
                        for (var i = 0; i < data.length; i++) {
                            newarray.push(data[i].date);
                        }

                        var sData = newarray.join();

                        $.post("/genReport", { tenant: tenant, data: sData })
                            .done(function (data) {
                                alert("Reponse from report generation" + data);
                            });
                    }
                }
            ],
            ajax: { url: '/tenantReport?tenant=' + tenant, dataSrc: "" },
            "rowCallback": function (row, data) {
                if ($.inArray(data.DT_RowId, selected) !== -1) {
                    $(row).addClass('selected');
                }
            }
        });
        table.buttons().container()
            .appendTo('#tenantReport .col-md-6:eq(0)');

    });

    $('#tenantReport tbody').on('click', 'tr', function () {
        var id = this.id;
        var index = $.inArray(id, selected);

        if (index === -1) {
            selected.push(id);
        } else {
            selected.splice(index, 1);
        }

        $(this).toggleClass('selected');
    });


    /*
     * Used to add spinners when processing a request
     */
    $("#deleteTenantFrm").on('submit', function () {
        $("#deleteTenant").remove()
        $("#deleteTenantBtn").prop("disabled", true)
        $("#deleteTenantBtn").html(
            `<span id="deleteTenant" iclass="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Deleting...`
        );
        $("#startTestBtn").prop("disabled", true)
        $("#stopTestBtn").prop("disabled", true)
        var form = $("#deleteTenantFrm")[0]; // You need to use standard javascript object here
        var formData = new FormData(form);
      
        // Call ajax for pass data to other place
        $.ajax({
            type: 'POST',
            enctype: 'multipart/form-data',
            url: '/delete-tenant',
            data: formData, // getting filed value in serialize form
            processData: false,
            contentType: false
        }).done(function (data) { // if getting done then call.
            $("#deleteTenantBtn").html(
                `<button type="submit" id="deleteTenantBtn" class="btn btn-primary">Delete Tenant</button>`
            )
            $("#startTestBtn").prop("disabled", false)
            $("#stopTestBtn").prop("disabled", false)
            alert(JSON.stringify(data))
            populate(data)

            })
            .fail(function () { // if fail then getting message
                $("#deleteTenantBtn").html(
                    `<button type="submit" id="deleteTenantBtn" class="btn btn-primary">Delete Tenant</button>`
                )
                $("#startTestBtn").prop("disabled", false)
                $("#stopTestBtn").prop("disabled", false)
                $("#GenericCreateTestMsg").empty()
                $("#GenericCreateTestMsg").append('<div class="alert alert-warning" role="alert">SERVER HAS CRASHED</div>')
            });

        // to prevent refreshing the whole page page
        return false;
    });

    $("#startTestFrm").on('submit', function () {
        $("#startTestBtn").prop("disabled", true)
        $("#startTestBtn").html(
            `<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Starting Test...`
        );
        $("#deleteTenantBtn").prop("disabled", true)
        $("#stopTestBtn").prop("disabled", true)
        var form = $("#startTestFrm")[0]; // You need to use standard javascript object here
        var formData = new FormData(form);
        formData.append('jmeter', $("#script-file-upload")[0].files[0]); 
        formData.append('data', $("#data-file-upload")[0].files[0]); 

        // Call ajax for pass data to other place
        $.ajax({
            type: 'POST',
            enctype: 'multipart/form-data',
            url: '/start-test',
            data: formData, // getting filed value in serialize form
            processData: false,
            contentType: false
        }).done(function (data) { // if getting done then call.
            $("#startTestBtn").html(
                `<button type="submit" id="startTestBtn" class="btn btn-primary">Run Test</button>`
            )
            $("#deleteTenantBtn").prop("disabled", false)
            $("#stopTestBtn").prop("disabled", false)
            alert(JSON.stringify(data))
            populate(data)

            })
            .fail(function () { // if fail then getting message
                $("#startTestBtn").html(
                    `<button type="submit" id="startTestBtn" class="btn btn-primary">Run Test</button>`
                )
                $("#deleteTenantBtn").prop("disabled", false)
                $("#stopTestBtn").prop("disabled", false)
                $("#GenericCreateTestMsg").empty()
                $("#GenericCreateTestMsg").append('<div class="alert alert-warning" role="alert">SERVER HAS CRASHED</div>')
            });

        // to prevent refreshing the whole page page
        return false;
    });

    $("#stopTestFrm").on('submit', function () {
        $("#stopTestBtn").prop("disabled", true)
        $("#stopTestBtn").html(
            `<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Stopping Test...`
        );
        $("#startTestBtn").prop("disabled", true)
        $("#deleteTenantBtn").prop("disabled", true)
    });

});

function populate(data){
    if (!_.isEmpty(data.RunningTests) && (data.RunningTests.length > 0)) {
        var form = '<div class="form-group">'+
        '<label for="context">Tennant</label>'+
        '<div id="RunningTests">'+
        '<div>'+
        '<select aria-label="Running Tests" class="form-control" name="stopcontext" id="stopcontext">'
        $.each(data.RunningTests, function( index, value ) {
            form = form.concat('<option value="'+value.Namespace+'">'+value.Namespace+'</option>');
          });
        
        var end = '</select>'+
          '<small id="tenantHelp" class="form-text text-muted">This is the tenant in which you want to stop the test for </small>'+
        '</div>'+
        '</div>'+
        '</div>'+
        '<button type="submit" id="stopTestBtn" class="btn btn-primary">Stop Test</button>'+
        '<div id="TennantNotStopped"></div>'+
       '<div id="TenantStopped"></div>';
         form = form.concat(end)
         $("#stopTestFrm").empty()
         $("#stopTestFrm").append(form)
    } else {
        $("#stopTestFrm").empty()
        $("#stopTestFrm").append('<div class="alert alert-warning" role="alert">No Tests are running</div>')
    }

    if (!_.isEmpty(data.AllTenants) && (data.AllTenants.length > 0) ){
        var form = '<div class="form-group">'+
        '<label for="context">Tennant</label>'+
        '<div id="RunningTests">'+
        '<div>'+
        '<select aria-label="Running Tests" class="form-control" name="TenantContext" id="TenantContext">';
        $.each(data.AllTenants, function( index, value ) {
            form = form.concat('<option value="'+value.Namespace+'">'+value.Namespace+'</option>');
          });
        
        var end = '</select>'+
          '<small id="tenantHelp" class="form-text text-muted">This is the tenant in which you want to stop the test for </small>'+
        '</div>'+
        '</div>'+
        '</div>'+
        '<button type="submit" id="deleteTenantBtn" class="btn btn-primary">Delete Tenant</button>'+
        '<div id="TennantNotDeleted"></div>'+
       '<div id="TenantDeleted"></div>';

         form = form.concat(end)
         $("#deleteTenantFrm").empty()
         $("#deleteTenantFrm").append(form)
    } else {
        $("#deleteTenantFrm").empty()
        $("#deleteTenantFrm").append('<div class="alert alert-warning" role="alert">No tenants have been created</div>')
    }
    /**
     * Check for validation errors
    */
     if (data.MissingTenant) {
        $("#MissingTenant").empty()
         $("#MissingTenant").append('<div class="alert alert-primary" role="alert"> You need to enter the tenant details</div>')
        //Add missing tenant
    } else {
        //Remove missing tenant
        $("#MissingTenant").empty()
    }

    //Check 
    if (data.MissingNumberOfNodes) {
        //Add missing number of nodes
        $("#MissingNumberOfNodes").empty()
        $("#MissingNumberOfNodes").append('<div class="alert alert-primary" role="alert"> You need to provide the number of node </div>')
    } else {
        $("#MissingNumberOfNodes").empty()
    }

    if (_.isEmpty(data.InvalidTenantName)) {
        $("#InvalidTenantName").empty()
    } else {
        $("#InvalidTenantName").empty()
        $("#InvalidTenantName").append('<div class="alert alert-warning" role="alert">Following can not be used as tenant names: '
        + data.InvalidTenantName +
        '</div>')   
    }

    if (_.isEmpty(data.GenericCreateTestMsg)) {
        $("#GenericCreateTestMsg").empty()
    } else {
        $("#GenericCreateTestMsg").empty()
        $("#GenericCreateTestMsg").append('<div class="alert alert-primary" role="alert"> Some thing did not go right: '
        + data.GenericCreateTestMsg +
        '</div>')
    }

    if (data.MissingJmeter) {
        $("#MissingJmeter").empty()
        $("#MissingJmeter").append('<div class="alert alert-primary" role="alert"> You need to provide the jmeter script to test</div>')
    } else {
        $("#MissingJmeter").empty()
    }


    if (data.MissingData) {
        $("#MissingData").empty()
        $("#MissingData").append('<div class="alert alert-primary" role="alert"> You need to provide the data file</div>')
    } else {
        $("#MissingData").empty()
    }

    if (_.isEmpty(data.TennantNotStopped)) {
        $("#TennantNotStopped").empty()
    }else {
        $("#TennantNotStopped").empty()
        $("#TennantNotStopped").append('<div class="alert alert-primary" role="alert"> <strong>Was not able to stop the test:'+data.TennantNotStopped+'</strong> </div>')
    }

    if (_.isEmpty(data.TenantStopped)) {
        $("#TenantStopped").empty()
    }else {
        $("#TenantStopped").empty()
        $("#TenantStopped").append('<div class="alert alert-success" role="alert"> <strong>Test were stopped for: '+data.TenantStopped+'</strong> </div>')
    }
    

    if (_.isEmpty(data.TennantNotDeleted)) {
        $("#TennantNotDeleted").empty()
    }else {
        $("#TennantNotDeleted").empty()
        $("#TennantNotDeleted").append('<div class="alert alert-primary" role="alert"> <strong> Tenant not deleted:'+data.TennantNotDeleted+'</strong> </div>')
    }

    if (_.isEmpty(data.TenantDeleted)) {
        $("#TenantDeleted").empty()
    }else {
        $("#TenantDeleted").empty()
        $("#TenantDeleted").append('<div class="alert alert-success" role="alert"> <strong>Tenant "'+data.TenantDeleted+'" has been deleted </strong> </div>')
    }
    

   



}
