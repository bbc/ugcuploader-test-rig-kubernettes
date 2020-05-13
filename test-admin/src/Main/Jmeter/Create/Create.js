import React, { Component } from "react";
import { post } from "axios";

import {
  Button,
  Container,
  Header,
  Form,
  Message,
  Modal,
} from "semantic-ui-react";
import SlaveConfig from "./SlaveConfig/SlaveConfig";
import "./Create.css";
const bandwidthOptions = [
  { key: "adsl", value: "adsl", text: "ADSL : 8 Mbit/s" },
  { key: "adsl2", value: "adsl2", text: "ADSL2 : 12 Mbit/s" },
  { key: "adsl2Plus", value: "adsl2Plus", text: "ADSL2+ : 24 Mbit/s" },
  {
    key: "ethernetLan",
    value: "ethernetLan",
    text: "Ethernet LAN ; 10 Mbit/s",
  },
  {
    key: "fastEthernet",
    value: "fastEthernet",
    text: "Fast Ethernet : 100 Mbit/s",
  },
  {
    key: "gigabitEthernet",
    value: "gigabitEthernet",
    text: "Gigabit Ethernet : 1 Gbit/s",
  },
  {
    key: "10gigabitEthernet",
    value: "10gigabitEthernet",
    text: "10 Gigabit Ethernet : 10 Gbit/s",
  },
  {
    key: "100gigabitEthernet",
    value: "100gigabitEthernet",
    text: "100 Gigabit Ethernet : 100 Gbit/s",
  },
  {
    key: "mobileDataEdge",
    value: "mobileDataEdge",
    text: "Mobile data EDGE : 384 kbit/s",
  },
  {
    key: "mobileDatacHspaPlus",
    value: "mobileDatacHspaPlus",
    text: "Mobile data HSPA+ : 21 Mbp/s",
  },
  {
    key: "mobileDataHspa",
    value: "mobileDataHspa",
    text: "Mobile data HSPA : 14,4 Mbp/s",
  },
  {
    key: "mobileDataDcHspaPlus",
    value: "mobileDataDcHspaPlus",
    text: "Mobile data DC-HSPA+ : 42 Mbps",
  },
  {
    key: "mobileDataLte",
    value: "mobileDataLte",
    text: "Mobile data LTE : 150 Mbp/s",
  },
  {
    key: "mobileDataGprs",
    value: "mobileDataGprs",
    text: "Mobile data GPRS : 171 kbit/s",
  },
  {
    key: "wifi80211a",
    value: "wifi80211a",
    text: "WIFI 802.11a/g : 54 Mbit/s",
  },
  { key: "wifi80211n", value: "wifi80211n", text: "WIFI 802.11n : 600 Mbit/s" },
];

class CreateTest extends Component {
  state = {
    open: false,
    teststart: false,
    bandwidthError: false,
    cpuError: false,
    jmeterscriptError: false,
    maxmetaError: false,
    ramError: false,
    slavesError: false,
    tenantError: false,
    testdataError: false,
    xmsError: false,
    xmxError: false,
    bandwidth: "",
    cpu: "",
    jmeterscript: "",
    maxmeta: "",
    ram: "",
    slaves: "",
    tenant: "",
    testdata: "",
    xms: "",
    xmx: "",
    formError: false,
  };

  handleChange = (e, item) => {
    if (item) {
      this.setState({ [item.name]: item.value });
    } else {
      if (e.target.name === "jmeterscript" || e.target.name === "testdata") {
        this.setState({ [e.target.name]: e.target.files[0] });
      } else {
        this.setState({ [e.target.name]: e.target.value });
      }
    }
  };
  handleSubmit = (e) => {
    e.preventDefault();
    console.log("state submit =", this.state);
    let error = false;
    if (this.state.bandwidth === "") {
      this.setState({ bandwidthError: true });
      error = true;
    } else {
      this.setState({ bandwidthError: false });
    }

    if (this.state.cpu === "") {
      this.setState({ cpuError: true });
      error = true;
    } else {
      this.setState({ cpuError: false });
    }

    if (this.state.jmeterscript === "") {
      this.setState({ jmeterscriptError: true });
      error = true;
    } else {
      this.setState({ jmeterscriptError: false });
    }

    if (this.state.maxmeta === "") {
      this.setState({ maxmetaError: true });
      error = true;
    } else {
      this.setState({ maxmetaError: false });
    }

    if (this.state.ram === "") {
      this.setState({ ramError: true });
      error = true;
    } else {
      this.setState({ ramError: false });
    }

    if (this.state.tenant === "") {
      this.setState({ tenantError: true });
      error = true;
    } else {
      this.setState({ tenantError: false });
    }
    if (this.state.testdata === "") {
      this.setState({ testdataError: true });
      error = true;
    } else {
      this.setState({ testdataError: false });
    }
    if (this.state.xms === "") {
      this.setState({ xmsError: true });
      error = true;
    } else {
      this.setState({ xmsError: false });
    }
    if (this.state.xmx === "") {
      this.setState({ xmxError: true });
      error = true;
    } else {
      this.setState({ xmxError: false });
    }
    if (this.state.slaves === "") {
      this.setState({ slavesError: true });
      error = true;
    } else {
      this.setState({ slavesError: false });
    }

    if (error === true) {
      this.setState({ formError: true });
    } else {
      this.setState({ formError: false });
      console.log("upload");
      this.fileUpload().then((response) => {
        console.log(response.data);
        this.setState({ teststart: true });
      });
    }
  };

  fileUpload(file) {
    const url = "/start-test";
    const formData = new FormData();
    formData.append("context", this.state.tenant);
    formData.append("numberOfNodes", this.state.slaves);
    formData.append("jmeter", this.state.jmeterscript);
    formData.append("data", this.state.testdata);
    formData.set("xmx", this.state.xmx);
    formData.set("xms", this.state.xms);
    formData.set("cpu", this.state.cpu);
    formData.set("ram", this.state.ram);
    formData.set("maxMetaspaceSize", this.state.maxmeta);
    const config = {
      headers: {
        "content-type": "multipart/form-data",
      },
    };
    return post(url, formData, config);
  }

  closeModal = () => {
    this.setState({ open: false });
    this.setState({ formError: false });
  };

  configureModal = (closeOnEscape, closeOnDimmerClick) => () => {
    this.setState({ open: true });
  };

  render() {
    const { open } = this.state;
    return (
      <Container className="CreateTest-Wrapper">
        <Container textAlign="center">
          <Header as="h1">Create Test</Header>
        </Container>

        <Form error={this.state.formError} onSubmit={this.handleSubmit}>
          {this.state.formError ? (
            <Message
              error
              header="Test Creation Error"
              content="Please make sure you have supplied all data"
            />
          ) : null}
          <Form.Field>
            <label>Tennant</label>
            <input
              name="tenant"
              onChange={this.handleChange}
              placeholder="Tennat"
            />
            <small>This is the tenant in which you want to run the tests</small>
          </Form.Field>
          <Form.Field>
            <label>Number of jmeter Slaves</label>
            <input
              name="slaves"
              onChange={this.handleChange}
              placeholder="Slaves"
            />
          </Form.Field>
          <Form.Group>
            <label>Settings for Jmeter slaves</label>
            <Button color="blue" onClick={this.configureModal(true, false)}>
              Configure Slaves
            </Button>
            <Modal style={{ height: "fit-content" }} open={open}>
              <Modal.Header>SlaveConfiguration</Modal.Header>
              <Modal.Content>
                <SlaveConfig create={this} />
              </Modal.Content>
              <Modal.Actions>
                <Button
                  positive
                  onClick={this.closeModal}
                  labelPosition="right"
                  icon="checkmark"
                  content="Done"
                />
              </Modal.Actions>
            </Modal>
          </Form.Group>
          <label>Select Bandwidth</label>
          <br />
          <Form.Select
            name="bandwidth"
            placeholder="select bandwidth"
            onChange={this.handleChange}
            options={bandwidthOptions}
          />
          <Form.Field>
            <label>Jmeter Test Script</label>
            <input
              name="jmeterscript"
              onChange={this.handleChange}
              type="file"
            />
            <small>This is the Jmeter script to upload</small>
          </Form.Field>
          <Form.Field>
            <label>Test Data</label>
            <input name="testdata" onChange={this.handleChange} type="file" />
            <small>This is the data file used by the test</small>
          </Form.Field>
          <Button color="blue" type="submit">
            Run Tests
          </Button>
        </Form>
        {this.state.teststart ? (
          <Message info header={this.state.tenant} content="Started Test" />
        ) : null}
      </Container>
    );
  }
}

export default CreateTest;
