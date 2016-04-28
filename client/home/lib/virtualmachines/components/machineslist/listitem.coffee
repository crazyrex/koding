kd = require 'kd'
React = require 'kd-react'
MachineDetails = require './machinedetails'
immutable = require 'immutable'

module.exports = class MachinesListItem extends React.Component

  @propTypes =
    machine              : React.PropTypes.instanceOf(immutable.Map).isRequired
    stack                : React.PropTypes.instanceOf(immutable.Map).isRequired
    shouldRenderDetails  : React.PropTypes.bool
    shouldRenderSpecs    : React.PropTypes.bool
    shouldRenderPower    : React.PropTypes.bool
    shouldRenderAlwaysOn : React.PropTypes.bool
    shouldRenderSharing  : React.PropTypes.bool
    onChangeAlwaysOn     : React.PropTypes.func
    onChangePowerStatus  : React.PropTypes.func
    onDetailOpen         : React.PropTypes.func
    onSharedWithUser     : React.PropTypes.func

  @defaultProps =
    shouldRenderDetails  : yes
    shouldRenderSpecs    : no
    shouldRenderPower    : yes
    shouldRenderAlwaysOn : no
    shouldRenderSharing  : no
    onChangeAlwaysOn     : kd.noop
    onChangePowerStatus  : kd.noop
    onDetailOpen         : kd.noop
    onSharedWithUser     : kd.noop


  constructor: (props) ->
    super props
    @state = {isDetailOpen: no}


  renderMachineDetails: ->
    return null  unless @props.shouldRenderDetails
    return null  unless @state.isDetailOpen

    <main className="MachinesListItem-machineDetails">
      <MachineDetails
        machine={@props.machine}
        shouldRenderSpecs={@props.shouldRenderSpecs}
        shouldRenderPower={@props.shouldRenderPower}
        shouldRenderAlwaysOn={@props.shouldRenderAlwaysOn}
        shouldRenderSharing={@props.shouldRenderSharing}
        onChangeAlwaysOn={@props.onChangeAlwaysOn}
        onChangePowerStatus={@props.onChangePowerStatus}
        onSharedWithUser={@props.onSharedWithUser}
      />
    </main>


  toggle: (event) ->

    isDetailOpen = not @state.isDetailOpen
    @setState { isDetailOpen }
    @props.onDetailOpen()  if isDetailOpen


  renderIpAddress: ->

    hasIp = @props.machine.get 'ipAddress'
    ip    = hasIp or '0.0.0.0'
    title = if hasIp then '' else 'Actual IP address will appear here when this VM is powered on.'

    <div title={title} className="MachinesListItem-hostName">{ip}</div>


  renderDetailToggle: ->

    return null  unless @props.shouldRenderDetails

    <div className="MachinesListItem-detailToggle#{if @state.isDetailOpen then ' expanded' else ''}">
      <button onClick={@bound 'toggle'}></button>
    </div>


  render: ->

    <div className="MachinesListItem#{if @state.isDetailOpen then ' expanded' else ''}">
      <header>
        <div
          className="MachinesListItem-machineLabel #{@props.machine.getIn ['status', 'state']}"
          onClick={@bound 'toggle'}>
          {@props.machine.get 'label'}
        </div>
        {@renderIpAddress()}
        <div className="MachinesListItem-stackLabel">
          <a href="#" className="HomeAppView--button primary">{@props.stack.get 'title'}</a>
        </div>
        {@renderDetailToggle()}
      </header>
      {@renderMachineDetails()}
    </div>



