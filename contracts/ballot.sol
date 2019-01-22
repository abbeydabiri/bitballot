pragma solidity >=0.4.22 <0.6.0;
contract Ballot {
    
    
    event Verified(address indexed _voter, bytes32 indexed _proposal);
    event NewProposal(bytes _name, uint _duration);
    event NewProposalPosition(bytes32 _title, bytes32 _proposal);
    event NewPositionCandidate(bytes32 _name, bytes _idHash, bytes32 _position, uint index);
    event Voted(address _voter, bytes32 _proposal, bytes32 _position, bytes32 _candidate);
    event NewVoter(address _voter, bytes32 _name, bytes _idHash);
    event Accredited(bytes32 _name, bytes32 _position);

    struct Voter {
        bytes32 name;
        bytes32 idHash;
        bool isUnique;
    }
    
    struct Eligibility {
        bool isVerified;
        bool isVoted;
        address voter;
    }

    struct Candidate {
        bytes32 position;
        bytes32 name;
        bytes idHash;
        bool isAccredited;
        uint voteCount;
    }
    
    struct Position {
        uint8 maxCandidate;
        bytes32 title;
        mapping(address => bytes) voted;
    }

    struct Proposal {
        uint dateAdded;
        uint duration;
        bytes32 name;
        bool isActive;
    }

    address initiator;
    Voter[] public aVoters;
    mapping(bytes32 => Position[]) public mProposalToPositions;
    mapping(bytes32 => Candidate[]) public mPositionToCanditate;
    mapping(bytes32 => Eligibility[]) public mProposalVoters;
    
    mapping(bytes32 => uint) internal mProposalToIndex;
    mapping(bytes32 => uint) internal mCandidateToIndex;
    mapping(bytes32 => uint) internal mPositionToIndex;
    mapping(address => uint) mEligibilityToIndex;
    mapping(address => uint) mVoterToIndex;
    Proposal[] public aProposals;

    constructor() public {
        initiator = msg.sender;
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
    
    function getEligibleVoters(bytes32 _proposal, uint _voterIndex) public view returns(bytes32 memory name, bytes memory idHash) {
        require(mProposalVoters[_proposal][_voterIndex].isVerified, "Not Eligible!");
        uint _voterIndex = mVoterToIndex[mProposalVoters[_proposal][_voterIndex].voter];
        return (aVoters[_voterIndex].name, aVoters[_voterIndex].idHash);
    }
    
    function addProposal (bytes32 _name, uint _dateAdded, uint _duration) public returns(uint) {
        uint index = aProposals.push(Proposal(_name, _dateAdded, _duration, false));
        mProposalToIndex[_name] = index - 1;
        emit NewProposal(_name, _duration);
        return index - 1;
    }
    
    function addPosition (bytes32 _title, uint8 _maxCandidate, bytes32 _proposal) public returns(uint) {
        uint index = mProposalToPositions[_proposal].push(Position(_title, _maxCandidate));
        mPositionToIndex[_title] = index - 1;
        emit NewProposalPosition(_title, _proposal);
        return index - 1;
    }
    
    function addCandidate (bytes32 _name, bytes32 _idHash, bytes32 _position, bytes32 _proposal) public returns(uint) {
        require(mProposalToPositions[_proposal][mPositionToIndex[_name]].maxCandidate >= mPositionToCanditate[_position].length, "Position maximum candidate exceeded!");
        uint index = mPositionToCanditate[_position].push(Candidate(_name, _idHash, _position, false, 0));
        mCandidateToIndex[_name] = index - 1;
        emit NewPositionCandidate(_name, _idHash, _position, index - 1);
        return index - 1;
    }
    
    function registerVoter(address _voter, bytes memory _name, bytes memory _idHash) public returns (uint) {
        if (aVoters.length > 0) {
            require(mVoterToIndex[_voter] != 0, "Voter already added!");
        }
        uint index = aVoters.push(Voter(_name, _idHash, true));
        mVoterToIndex[_voter] = index - 1;
        emit NewVoter(_voter, _name, _idHash);
        return index - 1;
    }

    function VerifyVoter(bytes32 _proposal, address _voter) public returns (bool){
        require(msg.sender == initiator, "Only the initiator of this ballot proposal can verify a voter");
        uint _voterIndex =  mVoterToIndex[_voter];
        require(aVoters[_voterIndex].isUnique, "Voter not found! Voter has to be added first");
        mProposalVoters[_proposal].push(Eligibility(_voter, true, false));
        emit Verified(_voter, _proposal);
        return true;
    }
    
    function accreditCandidate(bytes memory _name, bytes32 _position) public returns (bool){
        require(msg.sender == initiator, "Only the initiator of this ballot proposal can accredit a Candidate");
        uint index = mCandidateToIndex[_name];
        mPositionToCanditate[_position][index].isAccredited = true;
        emit Accredited(_name, _position);
        return true;
    }

    /// Give a single vote to proposal $(toProposal).
    function vote(bytes32 _proposal, bytes32 _position, bytes32 _candidate) public {
        require(aProposals[mProposalToIndex[_proposal]].dateAdded + aProposals[mProposalToIndex[_proposal]].duration == now, "Voting has ended for this proposal!");
        uint proposalIndex = mProposalToIndex[_proposal];
        uint candidateIndex = mCandidateToIndex[_candidate];
        uint positionIndex = mPositionToIndex[_position];
        uint eligibleIndex = mEligibilityToIndex[msg.sender];
        
        require(aProposals[proposalIndex].isActive, "This proposal is not active for votes!");
        require(mPositionToCanditate[_position][candidateIndex].isAccredited, "The candidate you want to vote for is not accredited!");
        require(mProposalVoters[_proposal][eligibleIndex].isVerified, "You are not eligible to vote on this proposal!");
        require(mProposalToPositions[_proposal][positionIndex].voted[msg.sender] == "", "You have voted in this position already");
        
        mPositionToCanditate[_position][candidateIndex].voteCount++;
        mProposalToPositions[_proposal][positionIndex].voted[msg.sender] = _candidate;
        emit Voted(msg.sender, _proposal, _position, _candidate);
        
    }
}
