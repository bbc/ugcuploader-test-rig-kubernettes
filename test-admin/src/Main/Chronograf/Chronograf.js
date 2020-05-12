import React, { Component } from "react";

import MyConsumer from "../../MyConsumer";
import "./Chronograf.css";

export class Chronograf extends Component {
  state = {
    items: {}
  };

  render() {
    return (
      <MyConsumer>
        {({ chronographUrl }) => (
    <object data={ chronographUrl} width="100%" height="100%" >
    </object>
        )}
      </MyConsumer>
    );
  }
}

export default Chronograf;
