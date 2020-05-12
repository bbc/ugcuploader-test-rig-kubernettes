import React, { Component } from "react";
import { post, get } from "axios";
import {
  Container,
  Button,
  Dropdown,
  Header,
  Segment,
} from "semantic-ui-react";
import "./TenantDeletion.css";
import _ from "lodash";

class TenantDeletion extends Component {
  state = { tennants: [] };

  componentDidMount() {
    this.fetchTestStatus();
  }

  componentDidUpdate() {
    //this.fetchTestStatus();
  }

  deleteTenant = () => {
    if (this.state.value) {
      const url = "/delete-tenant";
      const formData = new FormData();
      formData.set("TenantContext", this.state.value);

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
  fetchTestStatus = () => {
    get("/tenants").then((response) => {
      let tennants = response.data.AllTenants;
      console.log(tennants);

      let AllTenants = _.map(tennants, function (item) {
        return { key: item.Name, value: item.Name, text: item.Name };
      });
      this.setState({ tennants: AllTenants });
      //  this.setState({teststatus: status});
    });
  };
  render() {
    return (
      <Container className="Main-Wrapper">
        <Container textAlign="center">
          <Header as="h1">Tenant Deletion</Header>
        </Container>
        <Segment>
          <Dropdown
            onChange={this.handleChange}
            placeholder="Select Tenant to Delete"
            options={this.state.tennants}
          />
        </Segment>
        <Segment>
          <Button onClick={this.deleteTenant}> Delete Selected Tenant</Button>
        </Segment>
      </Container>
    );
  }
}

export default TenantDeletion;
