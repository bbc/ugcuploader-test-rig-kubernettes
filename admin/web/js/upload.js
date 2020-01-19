
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
                    text: 'Create Report for '+tenant,
                    action: function (e, dt, node, config) {

                        var data = table.rows({selected:  true}).data();
                        var newarray=[];       
                        for (var i=0; i < data.length ;i++){
                            newarray.push(data[i].date);
                         }
 
                        var sData = newarray.join();

                        $.post( "/genReport", { tenant: tenant, data: sData })
                        .done(function( data ) {
                            alert( "Reponse from report generation" + data );
                        });
                    }
                }
            ],
            ajax: { url: '/tenantReport?tenant=' + tenant, dataSrc: "" },
            "rowCallback": function( row, data ) {
                if ( $.inArray(data.DT_RowId, selected) !== -1 ) {
                    $(row).addClass('selected');
                }
            }
        });
        table.buttons().container()
        .appendTo( '#tenantReport .col-md-6:eq(0)' );

    });



    $('#tenantReport tbody').on('click', 'tr', function () {
        var id = this.id;
        var index = $.inArray(id, selected);
 
        if ( index === -1 ) {
            selected.push( id );
        } else {
            selected.splice( index, 1 );
        }
 
        $(this).toggleClass('selected');
    } );


    /*
     * Used to add spinners when processing a request
     */
    $("#deleteTenantFrm").on('submit', function () {
        $("#deleteTenant").remove()
        $("#deleteTenantBtn").prop("disabled", true)
        $("#deleteTenantBtn").html(
            `<span id="deleteTenant" class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Deleting...`
        );
        $("#startTestBtn").prop("disabled", true)
        $("#stopTestBtn").prop("disabled", true)
    });

    $("#startTestFrm").on('submit', function () {
        $("#startTestBtn").prop("disabled", true)
        $("#startTestBtn").html(
            `<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Starting Test...`
        );
        $("#deleteTenantBtn").prop("disabled", true)
        $("#stopTestBtn").prop("disabled", true)
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
