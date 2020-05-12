import React, { Component } from "react";
import { connect } from "react-redux";
import { Dropdown, Segment } from "semantic-ui-react";
import { reportFetchTenants } from "../../../../redux/actions";
import ReportList from "./ReportList/ReportList";
import "./GenerateGraph.css";
import _ from 'lodash';
import MyConsumer from "../../../../MyConsumer";

class GenerateGraph extends Component {

    constructor(props) {
        super(props)
        this.state ={}
    }

  componentDidMount() {
    this.props.reportFetchTenants();
  }

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
                  options={this.props.tenantList}
                  onChange ={(event, data) => {fetchReportsForTenant(data.value)}}
                  //onChange = {this.onChange}
                />
              </Segment>
              <Segment style={{ overflow: "auto", maxHeight: 200 }}>
                <ReportList />
              </Segment>
              </div>
        )}
      </MyConsumer>
    );
  }
}

/*
GenerateGraph.propTypes = {
  reportTenants: PropTypes.object.isRequired,
};
*/


function mapStateToProps(state, ownProps) {
    console.log("state=", state.reportFetchTenants.TenantList);
    const { reportFetchTenants } = state
    let tl = reportFetchTenants.TenantList;

    let con = _.map(tl, function(item) {
        return {key: item, value: item, text: item}
      });
      
    console.log("converted=", con);
    // component receives additionally:
    return { tenantList: con}
  }
  


const mapDispatchToProps = (dispatch) => ({
  reportFetchTenants: () => dispatch(reportFetchTenants()),
});

export default connect(mapStateToProps, mapDispatchToProps)(GenerateGraph);
