import React, { Component } from "react";

import BootstrapTable from "react-bootstrap-table-next";

import paginationFactory from "react-bootstrap-table2-paginator";
import MyConsumer from "../../../../../MyConsumer";
import "./ReportList.css";
const columns = [
  {
    dataField: "id",
    text: "id",
  },
  {
    dataField: "date",
    text: "date",
  },
];



export class ReportList extends Component {
  state = {
    items: {}
  };

 self = this;

 updateState = (row, isSelect, rowIndex, e) => {
    this.setState(state => {
        let items = this.state.items;
        console.log("items=",JSON.stringify(items));
        console.log("select=",isSelect);
        if (isSelect) {
            items[row.id] = row.date;
        } else {
             delete items[row.id];   
        }
        console.log("items-after="+JSON.stringify(items));
        return {
         items: items
        };
      });

      return true;

 }
 selectRowProp = {
    mode: 'checkbox', // single row selection
    hideSelectAll: true,
    clickToSelect: true,
     onSelect: this.updateState
  };
  componentDidMount() {}

  render() {
    return (
      <MyConsumer>
        {({ reports }) => (
          <BootstrapTable
          //  classes="reportlist"
            keyField="id"
            data={reports}
            selectRow={ this.selectRowProp }
            columns={columns}
            pagination={paginationFactory()}
          />
        )}
      </MyConsumer>
    );
  }
}

export default ReportList;
