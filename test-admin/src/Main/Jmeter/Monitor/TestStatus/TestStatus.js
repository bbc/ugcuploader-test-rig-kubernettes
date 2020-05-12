import React, { Component } from "react";
import BootstrapTable from "react-bootstrap-table-next";
import { get } from "axios";
import _ from "lodash";
import fetch from "isomorphic-fetch";
import { Container, Button, Header } from "semantic-ui-react";
import "./TestStatus.css";

const columns = [
  {
    dataField: "Tenant",
    text: "Tenant",
  },
  {
    dataField: "Started",
    text: "Started",
  },
  {
    dataField: "Errors",
    text: "Errors",
  },
];

class TestStatus extends Component {
  state = { teststatus: [] };

  fetchTestStatus = () => {
    get("/test-status").then((response) => {
      console.log("data", response.data);

      let deleted = response.data.BeingDeleted;
      let started = response.data.Started;
      console.log("started", started);
      console.log("deleted", deleted);

      if (started && deleted) {
        this.setState({ teststatus: started.concat(deleted) });
      } else if (started) {
        this.setState({ teststatus: started });
      } else if (deleted) {
        this.setState({ teststatus: deleted });
      } else {
        this.setState({ teststatus: [] });
      }
    });
  };
  render() {
    return (
      <Container className="Main-Wrapper">
        <Container textAlign="center">
          <Header as="h1">Test Status</Header>
        </Container>
        <BootstrapTable
          //  classes="reportlist"
          keyField="Tenant"
          data={this.state.teststatus}
          columns={columns}
        />
        <Button onClick={this.fetchTestStatus}> Fetch Test Status</Button>
      </Container>
    );
  }
}

export default TestStatus;
