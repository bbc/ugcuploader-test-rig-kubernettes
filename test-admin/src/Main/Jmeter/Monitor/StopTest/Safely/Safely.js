import React, { Component } from "react";
import { post, get } from "axios";
import {
  Container,
  Button,
  Dropdown,
  Segment,
} from "semantic-ui-react";
import "./Safely.css";
import _ from "lodash";

class Safely extends Component {
  state = { tennants: [] };

  componentDidMount() {
    this.fetchTenants();
  }

  stopTest = () => {
    if (this.state.value) {
      const url = "/stop-test";
      const formData = new FormData();
      formData.set("StopContext", this.state.value);

      const config = {
        headers: {
          "content-type": "multipart/form-data",
        },
      };
      post(url, formData, config).then((response) => {
        console.log(response);
      });
    }
  };

  handleChange = (e, { value }) => this.setState({ value });
  fetchTenants = () => {
    get("/tenants").then((response) => {
      let tennants = response.data.AllTenants;
      console.log(tennants);

      let AllTenants = _.map(tennants, function (item) {
        return { key: item.Namespace, value: item.Namespace, text: item.Namespace };
      });
      this.setState({ tennants: AllTenants });
    });
  };
  render() {
    return (
      <Container className="Main-Wrapper">
        <Segment>
          <Dropdown
            onChange={this.handleChange}
            placeholder="Select Tenant to Stop"
            options={this.state.tennants}
          />
        </Segment>
        <Segment>
          <Button color="blue" onClick={this.stopTest}> Stop Test Safely</Button>
        </Segment>
      </Container>
    );
  }
}

export default Safely;
