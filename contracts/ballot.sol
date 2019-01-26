pragma solidity >=0.4.22 <0.6.0;
contract Ballot {
    
    
    event Verified(uint indexed _voter, uint indexed _proposal,uint indexed _position);
    event NewProposal(bytes32 indexed _name, uint indexed _proposalId, uint _endDate);
    event NewPosition(uint indexed _proposal, uint indexed _position, bytes32 indexed _title);
    event NewCandidate(uint indexed _proposal, uint indexed _position, uint _candidateId, bytes32 indexed _name);
    event Voted(uint indexed _proposal, uint indexed _position, uint indexed _candidate, uint _voter, address _voterAddr);
    event NewVoter(address indexed _voter, uint _idHash);
    event Accredited(uint indexed _candidate, uint indexed _proposal, uint indexed _position);
    event ProposalActive(uint indexed _proposal);
    
    modifier onlyInitiator () {
        require(msg.sender == initiator, "Only the initiator of this ballot can perform this action");
        _;
    }
    
    address public initiator;
    
    struct Voter {
        uint voterId;
        bool isUnique;
        address voterAddr;
    }

    struct Vote {
        uint proposalId;
        uint positionId;
        uint candidateId;
        uint voterId;
    }
    
    struct EligibleVoter {
        bool isVerified;
        bool isVoted;
        uint positionId; // Users should be able to vote in multiple positions
    }

    struct Candidate {
        uint candidateId;
        uint positionId;
        uint proposalId;
        bool isAccredited;
        bool isUnique;
        address candidateAddr;
        bytes32 name;
    }

    struct Position {
        uint proposalId;
        uint positionId;
        uint8 maxCandidate;
        bytes32 title;
    }

    struct Proposal {
        uint proposalId;
        uint startDate;
        uint endDate;
        bytes32 name;
        bool isActive;
    }
    
    struct TrackIndex {
        uint index;
        bool isUnique;
    }
    
    mapping(uint => Position[]) public mPosition; // Maps proposal id to all it's positions
    mapping(uint => mapping(uint => Candidate[])) public mCandidate; // Maps proposal id to position id to candidates
    mapping(uint => EligibleVoter[]) public mProposalVoters;
    mapping(uint => mapping(uint => mapping(uint =>mapping(address => bool)))) internal votedCandidate;
    mapping(uint => mapping(uint => mapping(address => bool))) internal votedPosition;
    
    mapping(uint => TrackIndex) internal mProposalToIndex;
    mapping(uint => TrackIndex) internal mCandidateToIndex;
    mapping(uint => TrackIndex) internal mPositionToIndex;
    mapping(uint => uint) internal mEligibilityToIndex;
    mapping(address => uint) internal mVoterToIndex;
    
    Proposal[] public aProposals;
    Voter[] public aVoters;
    Vote[] public allVotes;

    constructor() public {
        initiator = msg.sender;
    }
    
    function votesCount() public view returns(uint) {
        return allVotes.length;
    }
    
    function votersCount() public view returns(uint) {
        return aVoters.length;
    }
    
    function eligibleVotersCount(uint _proposal) public view returns(uint) {
        return mProposalVoters[_proposal].length;
    }
    
    function positionsCount(uint _proposal) public view returns(uint) {
        return mPosition[_proposal].length;
    }
    
    function candidatesCount(uint _proposal, uint _position) public view returns(uint) {
        return mCandidate[_proposal][_position].length;
        
    }
    
    function getEligibleVoters(uint _proposal, address _voterAddr) public view returns(uint voterId, address voterAddr, uint positionId, bool isVerified, bool isVoted) {
        uint _voterIndex = mVoterToIndex[_voterAddr];
        uint _eligibleIndex = mEligibilityToIndex[_voterIndex];
        Voter memory _voter = aVoters[_voterIndex];
        EligibleVoter memory _eligible = mProposalVoters[_proposal][_eligibleIndex];
        require(_eligible.isVerified, "Not Eligible!");
        
        return (_voter.voterId, _voter.voterAddr, _eligible.positionId, _eligible.isVerified, _eligible.isVoted);
    }
    
    function addProposal (bytes32 _name, uint _proposalId, uint _endDate) public returns(uint) {
        require(!mProposalToIndex[_proposalId].isUnique, "Proposal already added!");
        require(_endDate >= now, "End date of proposal post me a date in the future");
        uint index = aProposals.push(Proposal( _proposalId, now, _endDate, _name, false));
        mProposalToIndex[_proposalId] = TrackIndex(index - 1, true);
        emit NewProposal(_name, _proposalId, _endDate);
        return index - 1;
    }
    
    function addPosition (bytes32 _title, uint _proposalId,  uint _positionId, uint8 _maxCandidate) public onlyInitiator() returns(uint positionId) {
        require(mProposalToIndex[_proposalId].isUnique, "Positions can only be added to existing prproposals!");
        require(!aProposals[mProposalToIndex[_proposalId].index].isActive && aProposals[mProposalToIndex[_proposalId].index].endDate >= now, "New positions cannot be added to an active or ended proposal");
        uint index = mPosition[_proposalId].push(Position(_proposalId, _positionId, _maxCandidate, _title));
        mPositionToIndex[_positionId] = TrackIndex(index - 1, true);
        emit NewPosition(_proposalId, _positionId, _title);
        return index - 1;
    }
    
    function addCandidate (address _candidateAddr, bytes32 _name, uint _positionId, uint _proposal, uint _candidateId) public onlyInitiator() returns(uint) {
        TrackIndex memory _IDProposal = mProposalToIndex[_proposal];
        TrackIndex memory _IDPosition = mPositionToIndex[_positionId];
        TrackIndex memory _IDCandidate = mCandidateToIndex[_candidateId];
        
        require(!_IDCandidate.isUnique, "Candidate has been added already!");
        require(_IDProposal.isUnique , "Candidate can only be added to an existing proposal");
        require(!aProposals[_IDProposal.index].isActive && aProposals[_IDProposal.index].endDate >= now , "Candidate cannot be added to an active or ended proposal");
        require(_IDPosition.isUnique , "Candidate can only be added to an existing position");
        require(mPosition[_proposal][_IDPosition.index].maxCandidate >= mCandidate[_proposal][_positionId].length, "Position maximum candidate exceeded!");
        
        uint index = mCandidate[_proposal][_positionId].push(Candidate(_candidateId, _positionId, _proposal, false, true, _candidateAddr, _name));
        mCandidateToIndex[_candidateId] = TrackIndex(index - 1, true);
        emit NewCandidate(_proposal, _positionId, _candidateId, _name);
        return index - 1;
    }
    
    function registerVoter(address _voter, uint _IDVoter) public returns (uint) {
        if (aVoters.length > 0) {
            require(mVoterToIndex[_voter] != 0, "Voter already added!");
        }
        uint index = aVoters.push(Voter(_IDVoter, true, _voter));
        mVoterToIndex[_voter] = index - 1;
        emit NewVoter(_voter, _IDVoter);
        return index - 1;
    }

    function VerifyVoter(uint _proposal, uint _positionId, uint _voterId, address _voterAddr) public onlyInitiator() returns (bool){
        require(!aProposals[mProposalToIndex[_proposal].index].isActive && aProposals[mProposalToIndex[_proposal].index].endDate >= now, "Voters cannot be verified for an active or ended proposal");
        uint _voterIndex =  mVoterToIndex[_voterAddr];
        require(aVoters[_voterIndex].isUnique, "Voter not found! Voter has to be added first");
        uint index = mProposalVoters[_proposal].push(EligibleVoter( true, false,_positionId));
        mEligibilityToIndex[_voterIndex] = index - 1;
        emit Verified(_voterId, _proposal, _positionId);
        return true;
    }
    
    function accreditCandidate(uint _candidateId, uint _positionId, uint _proposalId) public onlyInitiator() returns (bool){
        require(!aProposals[mProposalToIndex[_proposalId].index].isActive && aProposals[mProposalToIndex[_proposalId].index].endDate >= now, "Candidates cannot be accredited for an active or ended proposal");
        
        TrackIndex memory candidateIndex = mCandidateToIndex[_candidateId];
        mCandidate[_proposalId][_positionId][candidateIndex.index].isAccredited = true;
        emit Accredited(_candidateId, _proposalId, _positionId);
        return true;
    }
    
    function initiateVoting(uint _proposal) public onlyInitiator() returns (bool){
        TrackIndex memory _porposalId = mProposalToIndex[_proposal];
        aProposals[_porposalId.index].isActive = true;
        emit ProposalActive(_proposal);
        return true;
    }

    /// Give a single vote to proposal $(toProposal).
    function vote(uint _proposalId, uint _positionId, uint _candidateId, uint _voterId) public {
        TrackIndex memory _proposalIndex = mProposalToIndex[_proposalId];
        if(aProposals[_proposalIndex.index].endDate <= now) {
            aProposals[_proposalIndex.index].isActive = false;
        }
        
        require(aProposals[_proposalIndex.index].endDate > now, "Voting has ended for this proposal!");
        require(aProposals[_proposalIndex.index].isActive, "This proposal is not active for votes!");
        
        TrackIndex memory candidateIndex = mCandidateToIndex[_candidateId];
        uint _voterIndex =  mVoterToIndex[msg.sender];
        uint eligibleIndex = mEligibilityToIndex[_voterIndex];
        
        require(mCandidate[_proposalId][_positionId][candidateIndex.index].isAccredited, "The candidate you want to vote for is not accredited!");
        require(mProposalVoters[_proposalId][eligibleIndex].isVerified, "You are not eligible to vote on this proposal!");
        require(!votedCandidate[_proposalId][_positionId][_candidateId][msg.sender], "You have voted in this position already");
        require(!votedPosition[_proposalId][_positionId][msg.sender], "You have voted in this position already");

        allVotes.push(Vote(_proposalId,_positionId,_candidateId,_voterId));
        votedCandidate[_proposalId][_positionId][_candidateId][msg.sender] = true;
        votedPosition[_proposalId][_positionId][msg.sender] = true;
        mProposalVoters[_proposalId][eligibleIndex].isVoted = true;
        emit Voted(_proposalId, _positionId, _candidateId, _voterId, msg.sender);
        
    }
}