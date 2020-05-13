import React, { Component } from "react";
import BootstrapTable from "react-bootstrap-table-next";
import { get } from "axios";
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
      let deleted = response.data.BeingDeleted;
      let started = response.data.Started;
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
        <Button color="blue" onClick={this.fetchTestStatus}> Fetch Test Status</Button>
      </Container>
    );
  }
}

export default TestStatus;
