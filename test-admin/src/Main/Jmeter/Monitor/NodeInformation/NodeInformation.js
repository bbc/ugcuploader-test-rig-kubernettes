import React, {Component} from 'react';
import fetch from 'isomorphic-fetch'
import _ from "lodash";
import BootstrapTable from "react-bootstrap-table-next";
import { Button, Divider } from "semantic-ui-react";
import paginationFactory from "react-bootstrap-table2-paginator";

import './NodeInformation.css';
const columns = [
  {
    dataField: "Name",
    text: "Node",
  },
  {
    dataField: "InstanceID",
    text: "Instance ID",
  },
  {
    dataField: "Phase",
    text: "Phase",
  },
  {
    dataField: "NodeConditions",
    text: "NodeConditions",
  },
];

export class NodeInformation extends Component {
  state = {
    nodes: [],
  };

  fetchNodes = () => {
      fetch("/failing-nodes")
        .then((response) => {
            console.log("failg -nodes", response)
            if (response.status == 500) {
                return [{Name:"Something went wrong", InstanceID: "Error", Phase: "Error", NodeConditions: "Error"}]
            } else {
                return response.json();
            }
        })
        .then((json) => {
            this.setState({nodes: json});
        });
    
  };

  componentDidMount() {}

  render() {
    return (
      <div>
            <BootstrapTable
              keyField="InstanceID"
              data={this.state.nodes}
              columns={columns}
              pagination={paginationFactory()}
            />
            <Divider/>
            <Button color="blue" onClick={this.fetchNodes}>Fetch Node Details</Button>
          </div>
    );
  }
}

export default NodeInformation;
