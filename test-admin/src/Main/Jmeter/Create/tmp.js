
          <Form.Group>
          <label for="bandWidthSelection">Select the bandwidth</label>
          <select
            class="form-control"
            name="bandWidthSelection"
            id="bandWidthSelection"
          >
            <option value="adsl">ADSL : 8 Mbit/s</option>
            <option value="adsl2">ADSL2 : 12 Mbit/s</option>
            <option value="adsl2Plus">ADSL2+ : 24 Mbit/s</option>
            <option value="ethernetLan">Ethernet LAN ; 10 Mbit/s</option>
            <option value="fastEthernet">Fast Ethernet : 100 Mbit/s</option>
            <option value="gigabitEthernet">
              Gigabit Ethernet : 1 Gbit/s
            </option>
            <option value="10gigabitEthernet">
              10 Gigabit Ethernet : 10 Gbit/s
            </option>
            <option value="100gigabitEthernet">
              100 Gigabit Ethernet : 100 Gbit/s
            </option>
            <option value="mobileDataEdge">
              Mobile data EDGE : 384 kbit/s
            </option>
            <option value="mobileDataHspa">
              Mobile data HSPA : 14,4 Mbp/s
            </option>
            <option value="mobileDatacHspaPlus">
              Mobile data HSPA+ : 21 Mbp/s
            </option>
            <option value="mobileDataDcHspaPlus">
              Mobile data DC-HSPA+ : 42 Mbps
            </option>
            <option value="mobileDataLte">Mobile data LTE : 150 Mbp/s</option>
            <option value="mobileDataGprs">
              Mobile data GPRS : 171 kbit/s
            </option>
            <option value="wifi80211a">WIFI 802.11a/g : 54 Mbit/s</option>
            <option value="wifi80211n">WIFI 802.11n : 600 Mbit/s</option>
          </select>
        </Form.Group>
        <Form.Group>
          <label for="data-file-upload">Data</label>
          <input
            name="data"
            id="data-file-upload"
            type="file"
            class="file-upload"
            aria-describedby="jmeterHelp"
          />
          <small id="jmeterHelp" class="form-text text-muted">
            This is the data file used by the test
          </small>
        </Form.Group>
