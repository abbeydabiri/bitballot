pragma solidity >=0.4.22 <0.6.0;
contract Ballot {
    
    
    event Verified(address indexed _voter, bytes32 indexed _proposal);

    struct Voter {
        bytes32 name;
        bytes32 idHash;
        bool isVerified;
        bool isVoted;
        bool isUnique;
    }

    struct Candidate {
        bytes32 name;
        bytes32 idHash;
        bytes32 position;
        bool isAccredited;
        uint voteCount;
    }
    
    struct Position {
        bytes32 title;
        uint8 maxCandidate;
        mapping(address => bytes32) voted;
    }

    struct Proposal {
        bytes32 name;
        uint dateAdded;
        uint duration;
        bool isActive;
    }

    address initiator;
    mapping(address => Voter) public mVoters;
    mapping(address => Candidate) public mCandidates;
    mapping(bytes32 => Position[]) public mProposalToPositions;
    mapping(bytes32 => Candidate[]) public mPositionToCanditate;
    mapping(bytes32 => uint) internal mProposalToIndex;
    mapping(bytes32 => uint) internal mCandidateToIndex;
    mapping(bytes32 => uint) mPositionToIndex;
    Proposal[] public aProposals;

    constructor() public {
        initiator = msg.sender;
    }
    
    function addProposal (bytes32 _name, uint _dateAdded, uint _duration) public returns(uint) {
        uint index = aProposals.push(Proposal(_name, _dateAdded, _duration, false));
        mProposalToIndex[_name] = index - 1;
        return index - 1;
    }
    
    function addPosition (bytes32 _title, uint8 _maxCandidate, bytes32 _proposal) public returns(uint) {
        uint index = mProposalToPositions[_proposal].push(Position(_title, _maxCandidate));
        mPositionToIndex[_title] = index - 1;
        return index - 1;
    }
    
    function addCandidate (bytes32 _name, bytes32 _idHash, bytes32 _position) public returns(uint) {
        uint index = mPositionToCanditate[_position].push(Candidate(_name, _idHash, _position, false, 0));
        mCandidateToIndex[_name] = index - 1;
        return index - 1;
    }
    
    function registerVoter(address _voter, bytes32 _name, bytes32 _idHash, bytes32 _proposal) public {
        require(!mVoters[_voter].isUnique, "Voter already added!");
        mVoters[_voter].name = _name;
        mVoters[_voter].idHash = _idHash;
        mVoters[_voter].isUnique = true;
        emit Verified(_voter, _proposal);
    }

    function VerifyVoter(address _voter) public returns (bool){
        require(msg.sender == initiator, "Only the initiator of this ballot proposal can verify a voter");
        mVoters[_voter].isVerified = true;return true;
    }
    
    function accreditCandidate(bytes32 _name, bytes32 _position) public returns (bool){
        require(msg.sender == initiator, "Only the initiator of this ballot proposal can accredit a Candidate");
        uint index = mCandidateToIndex[_name];
        mPositionToCanditate[_position][index].isAccredited = true;
        return true;
    }

    /// Give a single vote to proposal $(toProposal).
    function vote(bytes32 _proposal, bytes32 _position, bytes32 _candidate) public {
        uint proposalIndex = mProposalToIndex[_proposal];
        uint candidateIndex = mCandidateToIndex[_candidate];
        uint positionIndex = mPositionToIndex[_position];
        
        require(aProposals[proposalIndex].isActive, "This proposal is not active for votes!");
        require(mPositionToCanditate[_position][candidateIndex].isAccredited, "The candidate you want to vote for is not accredited!");
        require(mVoters[msg.sender].isVerified, "You are not eligible to vote!");
        require(mProposalToPositions[_proposal][positionIndex].voted[msg.sender] == "", "You have voted in this position already");
        
        mPositionToCanditate[_position][candidateIndex].voteCount++;
        mProposalToPositions[_proposal][positionIndex].voted[msg.sender] = _candidate;
        
    }

    function winningProposal(bytes32 _proposal, bytes32 _position) public view returns (uint winningVoteCount, bytes32 winningCandidate) {
        uint positionIndex = mPositionToIndex[_position];
        
        require(mProposalToPositions[_proposal][positionIndex].title == _position, "Position not found in the proposal provided");
        uint256 winningVoteCount = 0;
        bytes32 winningCandidate;
        
        for (uint8 vote = 0; vote < mPositionToCanditate[_position].length; vote++)
            if (mPositionToCanditate[_position][vote].voteCount > winningVoteCount) {
                winningVoteCount = mPositionToCanditate[_position][vote].voteCount;
                winningCandidate = mPositionToCanditate[_position][vote].name;
            }
    }
}
