import React, { Component } from "react";
import BootstrapTable from "react-bootstrap-table-next";
import paginationFactory from "react-bootstrap-table2-paginator";
import MyConsumer from "../../../../../MyConsumer";
import { post } from "axios";
import { Button, Divider } from "semantic-ui-react";
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
    items: {},
  };

  updateState = (row, isSelect, rowIndex, e) => {
    this.setState((state) => {
      let items = this.state.items;
      console.log("items=", JSON.stringify(items));
      console.log("select=", isSelect);
      if (isSelect) {
        items[row.id] = row.date;
      } else {
        delete items[row.id];
      }
      console.log("items-after=" + JSON.stringify(items));
      return {
        items: items,
      };
    });

    return true;
  };

  generateReport = () => {
    console.log("generating report");

    if (this.state.items.length > 0) {

      const formData = new FormData();
      formData.set("data", this.state.items);
      formData.set("tenant", this.props.tenant);
      
      const config = {
        headers: {
          "content-type": "multipart/form-data",
        },
      };
      post("/genReport", formData, config).then((response) => {
        console.log(response);
      });

    }
  };
  selectRowProp = {
    mode: "checkbox", // single row selection
    hideSelectAll: true,
    clickToSelect: true,
    onSelect: this.updateState,
  };
  componentDidMount() {}

  render() {
    return (
      <MyConsumer>
        {({ reports }) => (
          <div>
            <BootstrapTable
              //  classes="reportlist"
              keyField="id"
              data={reports}
              selectRow={this.selectRowProp}
              columns={columns}
              pagination={paginationFactory()}
            />
            <Divider />
            <Button color="blue" onClick={this.generateReport}>
              Generate Report
            </Button>
          </div>
        )}
      </MyConsumer>
    );
  }
}

export default ReportList;
