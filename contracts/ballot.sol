pragma solidity >=0.4.22 <0.6.0;
contract Ballot {
    
    
    event Verified(address indexed _voter, bytes32 indexed _proposal);
    event NewProposal(bytes32 indexed _name, uint _endDate);
    event NewProposalPosition(bytes32 indexed _title, bytes32 indexed _proposal);
    event NewPositionCandidate(bytes32 indexed _name, bytes _idHash, bytes32 indexed _position, uint _candidateIndex);
    event Voted(address _voter, bytes32 indexed _proposal, bytes32 indexed _position, bytes32 indexed _candidate);
    event NewVoter(address indexed _voter, bytes32 _name, bytes _idHash);
    event Accredited(bytes32 indexed _name, bytes32 indexed _position);
    event ProposalActive(bytes32 indexed _proposal);
    
    modifier onlyInitiator (bytes32 _proposal) {
        require(msg.sender == initiator[_proposal], "Only the initiator of this ballot proposal can verify a voter");
        _;
    }
    
    address public owner;
    
    struct Voter {
        bytes32 name;
        bytes idHash;
        bool isUnique;
    }
    
    struct Eligibility {
        bool isVerified;
        bool isVoted;
        address voter;
    }

    struct Candidate {
        uint positionId;
        bytes32 name;
        bytes idHash;
        bool isAccredited;
        uint voteCount;
    }
    
    struct Position {
        uint proposalId;
        uint8 maxCandidate;
        bytes32 title;
        mapping(address => bytes32) voted;
    }

    struct Proposal {
        uint dateAdded;
        uint endDate;
        bytes32 name;
        bool isActive;
    }
    
    mapping(bytes32 => address) public initiator;
    mapping(bytes32 => Position[]) public mProposalToPositions;
    mapping(bytes32 => Candidate[]) public mPositionToCanditate;
    mapping(bytes32 => Eligibility[]) public mProposalVoters;
    
    mapping(bytes32 => uint) internal mProposalToIndex;
    mapping(bytes32 => uint) internal mCandidateToIndex;
    mapping(bytes32 => uint) internal mPositionToIndex;
    mapping(address => uint) internal mEligibilityToIndex;
    mapping(address => uint) internal mVoterToIndex;
    
    Proposal[] public aProposals;
    Voter[] public aVoters;

    constructor() public {
        owner = msg.sender;
    }
    
    function votersCount(bytes32 _proposal) public view returns(uint) {
        return aVoters.length;
    }
    
    function eligibleVotersCount(bytes32 _proposal) public view returns(uint) {
        return mProposalVoters[_proposal].length;
    }
    
    function proposalPositionCount(bytes32 _position) public view returns(uint) {
        return mPositionToCanditate[_position].length;
    }
    
    function positionToCanditateCount(bytes32 _proposal) public view returns(uint) {
        return mProposalToPositions[_proposal].length;
    }
    
    function getEligibleVoters(bytes32 _proposal, uint _voterIndex) public view returns(bytes32 name, bytes memory idHash) {
        require(mProposalVoters[_proposal][_voterIndex].isVerified, "Not Eligible!");
        return (aVoters[_voterIndex].name, aVoters[_voterIndex].idHash);
    }
    
    function addProposal (bytes32 _name, uint _endDate) public returns(uint) {
        require(mProposalToIndex[_name] > 0, "Proposal already added!");
        initiator[_name] = msg.sender;
        uint index = aProposals.push(Proposal( now, _endDate, _name, false));
        mProposalToIndex[_name] = index - 1;
        emit NewProposal(_name, _endDate);
        return index - 1;
    }
    
    function addPosition (bytes32 _title, uint8 _maxCandidate, bytes32 _proposal) public onlyInitiator(_proposal) returns(uint positionId) {
        uint _IDProposal = mProposalToIndex[_proposal];
        if(aProposals.length > 0){
            require(_IDProposal != 0, "Positions can ponly be added to valid proposals");
        }
        uint index = mProposalToPositions[_proposal].push(Position(_IDProposal, _maxCandidate, _title));
        mPositionToIndex[_title] = index - 1;
        emit NewProposalPosition(_proposal, _title);
        return index - 1;
    }
    
    function addCandidate (bytes32 _name, bytes memory _idHash, bytes32 _position, bytes32 _proposal) public onlyInitiator(_proposal) returns(uint) {
        uint _IDPosition = mPositionToIndex[_position];
        Proposal memory _proposalInstance = aProposals[mProposalToIndex[_proposal]];
        
        require(!_proposalInstance.isActive && _proposalInstance.endDate > now , "Candidate cannot be added to an active or ended proposal");
        require(mProposalToPositions[_proposal][_IDPosition].proposalId == mProposalToIndex[_proposal] , "Candidate cannot be added to a position that does not exist!");
        require(mProposalToPositions[_proposal][_IDPosition].maxCandidate >= mPositionToCanditate[_position].length, "Position maximum candidate exceeded!");
        
        uint index = mPositionToCanditate[_position].push(Candidate(_IDPosition, _name, _idHash, false, 0));
        mCandidateToIndex[_name] = index - 1;
        emit NewPositionCandidate(_name, _idHash, _position, index - 1);
        return index - 1;
    }
    
    function registerVoter(address _voter, bytes32 _name, bytes memory _idHash) public returns (uint) {
        if (aVoters.length > 0) {
            require(mVoterToIndex[_voter] != 0, "Voter already added!");
        }
        uint index = aVoters.push(Voter(_name, _idHash, true));
        mVoterToIndex[_voter] = index - 1;
        emit NewVoter(_voter, _name, _idHash);
        return index - 1;
    }

    function VerifyVoter(bytes32 _proposal, address _voter) public onlyInitiator(_proposal) returns (bool){
        uint _voterIndex =  mVoterToIndex[_voter];
        require(aVoters[_voterIndex].isUnique, "Voter not found! Voter has to be added first");
        mProposalVoters[_proposal].push(Eligibility( true, false, _voter));
        emit Verified(_voter, _proposal);
        return true;
    }
    
    function accreditCandidate(bytes32 _name, bytes32 _position, bytes32 _proposal) public onlyInitiator(_proposal) returns (bool){
        uint index = mCandidateToIndex[_name];
        mPositionToCanditate[_position][index].isAccredited = true;
        emit Accredited(_name, _position);
        return true;
    }
    
    function initiateVoting(bytes32 _proposal) public onlyInitiator(_proposal) returns (bool){
        uint _porposalId = mProposalToIndex[_proposal];
        aProposals[_porposalId].isActive = true;
        emit ProposalActive(_proposal);
        return true;
    }

    /// Give a single vote to proposal $(toProposal).
    function vote(bytes32 _proposal, bytes32 _position, bytes32 _candidate) public {
        uint _proposalIndex = mProposalToIndex[_proposal];
        if(aProposals[_proposalIndex].endDate >= now) {
            aProposals[_proposalIndex].isActive = false;
        }
        
        require(aProposals[_proposalIndex].endDate > now, "Voting has ended for this proposal!");
        require(aProposals[_proposalIndex].isActive, "This proposal is not active for votes!");
        
        uint candidateIndex = mCandidateToIndex[_candidate];
        uint positionIndex = mPositionToIndex[_position];
        uint eligibleIndex = mEligibilityToIndex[msg.sender];
        
        require(mPositionToCanditate[_position][candidateIndex].isAccredited, "The candidate you want to vote for is not accredited!");
        require(mProposalVoters[_proposal][eligibleIndex].isVerified, "You are not eligible to vote on this proposal!");
        require(mProposalToPositions[_proposal][positionIndex].voted[msg.sender].length == 0, "You have voted in this position already");
        
        mPositionToCanditate[_position][candidateIndex].voteCount++;
        mProposalToPositions[_proposal][positionIndex].voted[msg.sender] = _candidate;
        emit Voted(msg.sender, _proposal, _position, _candidate);
        
    }
}
