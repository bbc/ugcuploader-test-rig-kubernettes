import React, { Component } from "react";
import { Dropdown, Segment } from "semantic-ui-react";
import { get } from "axios";
import "./SlaveLogs.css";
import _ from 'lodash';
import MyConsumer from "../../../../MyConsumer";

class SlaveLogs extends Component {

    constructor(props) {
        super(props)
        this.state ={tennants: []}
    }

  componentDidMount() {
    this.fetchTestStatus();
  }

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

  // onChange = (event, data)=> {
  //   console.log("event =======>",event);
  //   console.log(data);
  // } 
  render() {
    return (
      <MyConsumer>
        {({ fetchReportsForTenant }) => (
            <div>
              <Segment>
                <Dropdown
                  placeholder="Select tennet"
                  options={this.state.tennants}
                  //onChange ={(event, data) => {fetchReportsForTenant(data.value)}}
                  //onChange = {this.onChange}
                />
              </Segment>
              </div>
        )}
      </MyConsumer>
    );
  }
}



export default SlaveLogs;
