import React, { Component } from "react";

import MyConsumer from "../../MyConsumer";
import "./Report.css";

export class Report extends Component {
  state = {
    items: {}
  };

  render() {
    return (
      <MyConsumer>
        {({ graphsUrl }) => (
    <object data={ graphsUrl} width="100%" height="100%" >
    </object>
        )}
      </MyConsumer>
    );
  }
}

export default Report;
