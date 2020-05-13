import React, { Component } from "react";

import MyConsumer from "../../MyConsumer";
import "./Report.css";
import { Segment } from "semantic-ui-react";

export class Report extends Component {
  state = {
    items: {},
  };

  render() {
    return (
      <MyConsumer>
        {({ graphsUrl }) => (
          <Segment className="Report">
            <object data={graphsUrl} width="100%" height="100%"></object>
          </Segment>
        )}
      </MyConsumer>
    );
  }
}

export default Report;
