import React, { Component } from "react";
import { Container, Dropdown, Grid, Header, Segment} from "semantic-ui-react";
import "./SlaveConfig.css";


const memoryOptions = [
    { key: "1", value: "1", text: "1 Gigabyte"},
    { key: "2", value: "2", text: "2 Gigabyte"},
    { key: "3", value: "3", text: "3 Gigabyte"},
    { key: "4", value: "4", text: "4 Gigabyte"},
    { key: "5", value: "5", text: "5 Gigabyte"},
    { key: "6", value: "6", text: "6 Gigabyte"}
];

const metaSpace = [
    {key: "256", value: "256", text: "256mb"},
    {key: "512", value: "512", text: "512mb"},
    {key: "768", value: "768", text: "768mb"},
    {key: "1024", value: "1024", text: "1024mb"},
    {key: "1280", value: "1280", text: "1280mb"},
    {key: "1536", value: "1536", text: "1536mb"},
];

const cpus = [
    { key: "1", value: "1", text: "1 CPU"},
    { key: "2", value: "2", text: "2 CPU"},
    { key: "3", value: "3", text: "3 CPU"},
    { key: "4", value: "4", text: "4 CPU"}
];

const ram = [
    { key: "1", value: "1", text: "1 Gigabyte"},
    { key: "2", value: "2", text: "2 Gigabyte"},
    { key: "3", value: "3", text: "3 Gigabyte"},
    { key: "4", value: "4", text: "4 Gigabyte"},
    { key: "5", value: "5", text: "5 Gigabyte"},
    { key: "6", value: "6", text: "6 Gigabyte"},
    { key: "7", value: "7", text: "7 Gigabyte"},
    { key: "8", value: "8", text: "8 Gigabyte"},
    { key: "9", value: "9", text: "9 Gigabyte"},
    { key: "10", value: "10", text: "10 Gigabyte"},
    { key: "11", value: "11", text: "11 Gigabyte"},
    { key: "12", value: "12", text: "12 Gigabyte"},
    { key: "13", value: "13", text: "13 Gigabyte"},
    { key: "14", value: "14", text: "14 Gigabyte"}

]

class SlaveConfig extends Component {
    state = {};

    handleChange = (e, item) => {
        alert(JSON.stringify(item.value));

        this.props.create.check(item.name, item.value);
    }
    render() {
      return (
        <Container className="Jmeter-Wrapper">
        <Grid divided="vertically">
          <Grid.Row columns={2}>
            <Grid.Column>
            <Header as="h1">JVM Settings</Header>
            <Segment>
                <label>Initial Memory allocation (Xms):</label>
            <Dropdown  name="xms" onChange={this.props.create.handleChange} placeholder="Select Memory allocation (Xms)" options={memoryOptions} />
            </Segment>
           <Segment>
           <label>Maximum Memory allocation (Xmx):</label>
            <Dropdown name="xmx" onChange={this.props.create.handleChange} placeholder="Select Maximum Memory allocation (Xmx)" options={memoryOptions} />
            </Segment>
           <Segment> 
           <label>Control Garbage Collection  (MaxMetaspaceSize):</label>
            <Dropdown name="maxmeta" onChange={this.props.create.handleChange} placeholder="Select (MaxMetaspaceSize)" options={metaSpace} />
            </Segment>
            </Grid.Column>
            <Grid.Column>
            <Header as="h1">Node Settings</Header>
            <Segment>
            <label>Allocate cpu:</label>
            <Dropdown name="cpu" onChange={this.props.create.handleChange} placeholder="Select Cpu" options={cpus} />
            </Segment>
            <Segment>
            <label>Allocate Bandwidth:</label>
            <Dropdown name="ram" onChange={this.props.create.handleChange} placeholder="Select Bandwidth" options={ram} />
            </Segment>
            </Grid.Column>
          </Grid.Row>
        </Grid>
      </Container>
   
        );
    }
  }
  
  export default SlaveConfig;